package graphrouter

import (
	"fmt"
	"net"
	"net/http"
	"time"

	gr "github.com/sdqri/sequined/internal/graphgenerator"
	"github.com/sdqri/sequined/internal/hyperrenderer"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
	obs "github.com/sdqri/sequined/internal/observer"
)

type GraphServeMux struct {
	*http.ServeMux
	Observer *obs.Observer
}

type GraphServeMuxOption func(*GraphServeMux)

func New(opts ...GraphServeMuxOption) *GraphServeMux {
	graphServeMux := GraphServeMux{
		ServeMux: http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(&graphServeMux)
	}

	return &graphServeMux
}

func WithObserver(observer *obs.Observer) GraphServeMuxOption {
	return func(mux *GraphServeMux) {
		mux.Observer = observer
	}
}

func (mux *GraphServeMux) ServeGraph(
	root hr.HyperRenderer, updateChan chan gr.UpdateMessage, errChan chan error) {
	if mux.Observer != nil {
		hyperrenderer.Traverse(root, func(hr hr.HyperRenderer) bool {
			currentPage, ok := hr.(*hyperrenderer.Webpage)
			if !ok {
				// err := ErrUnexpectedNodeType
				//TOOD: handle error
				return true
			}

			mux.Observer.LogNode(obs.NodeLog{
				ID:        obs.NodeID(currentPage.GetID()),
				CreatedAt: time.Now(),
			})

			return false
		})
	}

	routeMap := hr.CreatePathMap(root)
	go func() {
		for updateMsg := range updateChan {
			routeMap = hr.CreatePathMap(root)
			if mux.Observer != nil {
				switch updateMsg.Type {
				case gr.UpdateTypeCreate:
					mux.Observer.LogNode(obs.NodeLog{
						ID:        obs.NodeID(updateMsg.Webpage.GetID()),
						CreatedAt: time.Now().UTC(),
						DeletedAt: nil,
					})
				case gr.UpdateTypeDelete:
					now := time.Now().UTC()
					if logNode, ok := mux.Observer.NodeLogMap[obs.NodeID(updateMsg.Webpage.GetID())]; ok {
						logNode.DeletedAt = &now
					}
				}
			}
		}
	}()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return
		}

		page, ok := routeMap[r.URL.Path]
		if ok {
			err := page.Render(w)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			http.NotFound(w, r)
		}

		if mux.Observer != nil {
			mux.Observer.LogVisit(obs.VisitLog{
				RemoteAddr: obs.IPAddr(ip),
				NodeID:     obs.NodeID(page.GetID()),
				VisitedAt:  time.Now().UTC(),
			})
		}
	})

}

func (mux *GraphServeMux) ActivateObserver() {
	if mux.Observer == nil {
		//TODO: should return error
	}
	mux.HandleFunc("/observer/age", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if ip := query.Get("ip"); ip != "" {
			result := fmt.Sprintf("freshness = %v", mux.Observer.GetFreshness(ip, time.Now()))
			fmt.Fprintf(w, result)
		}
	})
}
