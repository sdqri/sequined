package graphgenerator

import (
	"errors"
	"time"

	hr "github.com/sdqri/sequined/internal/hyperrenderer"
)

type SelectorFunc func(probabilities []float64) (int, error)

type GraphGenerator struct {
	Root                   *hr.Webpage
	PreferentialAttachment float64
	SelectorFunc
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
		probability := (float64(linkCount)/float64(totalHubsLinksCount))*
			float64(gg.PreferentialAttachment) +
			(1-gg.PreferentialAttachment)*(1/float64(totalHubsCount))
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
	hubsCount := 0
	AuthsCount := 0

	var err error = nil
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			err = ErrUnexpectedNodeType
			return true
		}
		if currentPage.Type == hr.WebpageTypeHub {
			hubsCount++
		} else if currentPage.Type == hr.WebpageTypeAuthority {
			AuthsCount++
		}

		return false
	})

	if err != nil {
		return err
	}

	for hubsCount < maxHubCount {
		gg.CreateHubPage()
		hubsCount++
	}

	for AuthsCount < maxAuthCount {
		gg.CreateAuthorityPage()
		AuthsCount++
	}
	return nil
}

func (gg *GraphGenerator) StartGraphEvolution(
	maxHubCount, maxAuthCount int,
	authCreationRate float64, hubCreationRate float64,
) (chan UpdateMessage, chan error) {
	// TODO: Determine optimal buffer size based on expected usage patterns
	updateChan := make(chan UpdateMessage, maxHubCount+maxAuthCount)
	errChan := make(chan error, maxHubCount+maxAuthCount)

	// Since the rate is specified in pages per hour,
	authCreationInterval := time.Hour / time.Duration(authCreationRate)
	hubCreationInterval := time.Hour / time.Duration(hubCreationRate)

	// Count existing hub and authority pages
	hubsCount := 0
	authsCount := 0
	hr.Traverse(gg.Root, func(currentRenderer hr.HyperRenderer) bool {
		currentPage, ok := currentRenderer.(*hr.Webpage)
		if !ok {
			errChan <- ErrUnexpectedNodeType
			return false
		}
		if currentPage.Type == hr.WebpageTypeHub {
			hubsCount++
		} else if currentPage.Type == hr.WebpageTypeAuthority {
			authsCount++
		}
		return false
	})

	// TODO: What happens to other goroutine if one of them errors, sending to chan will panic!
	// Create hub pages Asynchronously
	go func() {
		for hubsCount < maxHubCount {
			webpage, err := gg.CreateHubPage()
			if err != nil {
				errChan <- err
				break
			}
			hubsCount++
			updateChan <- UpdateMessage{
				Type:    UpdateTypeCreate,
				Webpage: webpage,
			}
			time.Sleep(hubCreationInterval)
		}

		close(updateChan)
		close(errChan)
	}()

	// Create auth pages Asynchronously
	go func() {
		for authsCount < maxAuthCount {
			webpage, err := gg.CreateAuthorityPage()
			if err != nil {
				errChan <- err
				break
			}
			authsCount++
			updateChan <- UpdateMessage{
				Type:    UpdateTypeCreate,
				Webpage: webpage,
			}
			time.Sleep(authCreationInterval)
		}

		close(updateChan)
		close(errChan)
	}()

	return updateChan, errChan
}
