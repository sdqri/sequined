package graphgenerator

import (
	"errors"
	"sync"
	"time"

	hr "github.com/sdqri/sequined/internal/hyperrenderer"
)

var (
	ErrMaxHubOrAuthCountAlreadyExceeded error = errors.New("maxHubCount or maxAuthCount is already exceeded")
)

type SelectorFunc func(probabilities []float64) (int, error)

type GraphGenerator struct {
	Root                   *hr.Webpage
	PreferentialAttachment float64
	SelectorFunc
	mu sync.Mutex
}

func New(root *hr.Webpage, preferentialAttachment float64) *GraphGenerator {
	return &GraphGenerator{
		Root:                   root,
		PreferentialAttachment: preferentialAttachment,
		SelectorFunc:           SelectByProbability,
	}
}

type UpdateType string

const (
	UpdateTypeCreate UpdateType = "create"
	UpdateTypeDelete UpdateType = "delete"
)

type UpdateMessage struct {
	Type    UpdateType
	Webpage *hr.Webpage
}

var (
	ErrUnexpectedNodeType error = errors.New("unexpected node type while traversing graph")
)

func (gg *GraphGenerator) CreateHubPage() (*hr.Webpage, error) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	hubsMap := make(map[*hr.Webpage]bool)
	totalHtoHLinksCount := 0

	var err error = nil
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			err = ErrUnexpectedNodeType
			return true
		}

		if currentPage.Type == hr.WebpageTypeHub {
			totalHtoHLinksCount += len(currentPage.Links)
			hubsMap[currentPage] = true
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	totalHubsCount := len(hubsMap)

	if totalHubsCount == 0 {
		webpage := gg.Root.AddChild(hr.WebpageTypeHub)
		return webpage, nil
	}

	probabilities := make([]float64, 0, totalHubsCount)
	hubNodes := make([]*hr.Webpage, 0, totalHubsCount)
	for node := range hubsMap {
		htohLinkCount := len(node.Links)
		probability := float64(1) / float64(totalHubsCount)
		if totalHtoHLinksCount != 0 {
			probability = (float64(htohLinkCount)/float64(totalHtoHLinksCount))*
				float64(gg.PreferentialAttachment) +
				(1-gg.PreferentialAttachment)*(1/float64(totalHubsCount))
		}
		probabilities = append(probabilities, probability)
		hubNodes = append(hubNodes, node)
	}

	hubIndex, err := gg.SelectorFunc(probabilities)
	if err != nil {
		return nil, err
	}

	webpage := hubNodes[hubIndex].AddChild(hr.WebpageTypeHub)
	return webpage, nil
}

func (gg *GraphGenerator) CreateAuthorityPage() (*hr.Webpage, error) {
	gg.mu.Lock()
	defer gg.mu.Unlock()

	hubsMap := make(map[*hr.Webpage]bool)
	totalHubsLinksCount := 0

	var err error = nil
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			err = ErrUnexpectedNodeType
			return true
		}

		if currentPage.Type == hr.WebpageTypeHub {
			totalHubsLinksCount += len(currentPage.Links)
			hubsMap[currentPage] = true
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	totalHubsCount := len(hubsMap)

	probabilities := make([]float64, 0, totalHubsCount)
	hubNodes := make([]*hr.Webpage, 0, totalHubsCount)

	for node := range hubsMap {
		linkCount := len(node.Links)
		probability := float64(1) / float64(totalHubsCount)
		if totalHubsLinksCount != 0 {
			probability = (float64(linkCount)/float64(totalHubsLinksCount))*
				float64(gg.PreferentialAttachment) +
				(1-gg.PreferentialAttachment)*(1/float64(totalHubsCount))
		}
		probabilities = append(probabilities, probability)
		hubNodes = append(hubNodes, node)
	}

	hubIndex, err := gg.SelectorFunc(probabilities)
	if err != nil {
		return nil, err
	}

	webpage := hubNodes[hubIndex].AddChild(hr.WebpageTypeAuthority)
	return webpage, nil
}

func (gg *GraphGenerator) Generate(maxHubCount, maxAuthCount int) error {
	hubCount := 0
	authCount := 0

	var err error = nil
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			err = ErrUnexpectedNodeType
			return true
		}
		if currentPage.Type == hr.WebpageTypeHub {
			hubCount++
		} else if currentPage.Type == hr.WebpageTypeAuthority {
			authCount++
		}

		return false
	})

	if err != nil {
		return err
	}

	if hubCount > maxHubCount || authCount > maxAuthCount {
		return ErrMaxHubOrAuthCountAlreadyExceeded
	}

	for hubCount < maxHubCount {
		gg.CreateHubPage()
		hubCount++
	}
	for authCount < maxAuthCount {
		gg.CreateAuthorityPage()
		authCount++
	}
	return nil
}

func (gg *GraphGenerator) StartGraphEvolution(
	maxHubCount, maxAuthCount int,
	authCreationRate float64, hubCreationRate float64,
) (chan UpdateMessage, chan error, error) {
	// Since the rate is specified in pages per hour,
	authCreationInterval := time.Hour / time.Duration(authCreationRate)
	hubCreationInterval := time.Hour / time.Duration(hubCreationRate)

	// Count existing hub and authority pages
	hubCount := 0
	authCount := 0
	var err error
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			err = ErrUnexpectedNodeType
			return false
		}
		if currentPage.Type == hr.WebpageTypeHub {
			hubCount++
		} else if currentPage.Type == hr.WebpageTypeAuthority {
			authCount++
		}
		return false
	})

	if err != nil {
		return nil, nil, err
	}

	if hubCount > maxHubCount || authCount > maxAuthCount {
		return nil, nil, ErrMaxHubOrAuthCountAlreadyExceeded
	}

	actionsCount := (maxHubCount - hubCount) + (maxAuthCount - authCount)
	updateChan := make(chan UpdateMessage, actionsCount)
	errChan := make(chan error, actionsCount)
	GeneratorErrorChan := make(chan struct{}, 0)
	GeneratorDoneChan := make(chan struct{}, 0)

	// Creates hub pages Asynchronously
	retreatChan := make(chan struct{}, 0)
	go func() {
		ticker := time.NewTicker(hubCreationInterval)
		defer ticker.Stop()
	outerLoop:
		for {
			select {
			case <-ticker.C:
				if hubCount < maxHubCount {
					webpage, err := gg.CreateHubPage()
					if err != nil {
						errChan <- err
						GeneratorErrorChan <- struct{}{}
						retreatChan <- struct{}{}
						return
					}
					hubCount++
					updateChan <- UpdateMessage{
						Type:    UpdateTypeCreate,
						Webpage: webpage,
					}
				} else {
					break outerLoop
				}
			case <-retreatChan:
				GeneratorErrorChan <- struct{}{}
				return
			}
		}
		GeneratorDoneChan <- struct{}{}
	}()

	go func() {
		ticker := time.NewTicker(authCreationInterval)
		defer ticker.Stop()
	outerLoop:
		for {
			select {
			case <-ticker.C:
				if authCount < maxAuthCount {
					webpage, err := gg.CreateAuthorityPage()
					if err != nil {
						errChan <- err
						GeneratorErrorChan <- struct{}{}
						retreatChan <- struct{}{}
						return
					}
					authCount++
					updateChan <- UpdateMessage{
						Type:    UpdateTypeCreate,
						Webpage: webpage,
					}
				} else {
					break outerLoop
				}
			case <-retreatChan:
				GeneratorErrorChan <- struct{}{}
				return
			}
		}
		GeneratorDoneChan <- struct{}{}
	}()

	go func() {
		countDone := 0
		countRetreated := 0
	outerLoop:
		for {
			select {
			case <-GeneratorErrorChan:
				countRetreated++
				if countRetreated == 2 {
					break outerLoop
				}
			case <-GeneratorDoneChan:
				countDone++
				if countDone == 2 {
					break outerLoop
				}
			}
		}
		close(updateChan)
		close(errChan)
		close(GeneratorDoneChan)
		close(GeneratorErrorChan)
		close(retreatChan)
	}()

	return updateChan, errChan, nil
}
