package graphmultiplexer

import (
	"fmt"
	"net/http"
	"time"

	gr "github.com/sdqri/sequined/internal/graphgenerator"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
	obs "github.com/sdqri/sequined/internal/observer"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type GraphMux struct {
	Root     *hr.Webpage
	RouteMap map[string]hr.HyperRenderer
	Observer *obs.Observer

	*http.ServeMux
	middlewareChain []Middleware
}

type GraphMuxOption func(*GraphMux)

func New(root *hr.Webpage, opts ...GraphMuxOption) (*GraphMux, error) {
	routeMap := hr.CreatePathMap(root)

	mux := GraphMux{
		Root:     root,
		RouteMap: routeMap,

		ServeMux: http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(&mux)
	}

	if mux.Observer != nil {
		var err error
		hr.Traverse(mux.Root, func(node hr.HyperRenderer) bool {
			currentPage, ok := node.(*hr.Webpage)
			if !ok {
				err = gr.ErrUnexpectedNodeType
				return true
			}
			mux.logNodeCreation(currentPage)
			return false
		})
		if err != nil {
			return nil, err
		}
	}

	GraphHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.HandleGraphHttpRequest(w, r)
	})

	for _, mw := range mux.middlewareChain {
		GraphHandlerFunc = mw(GraphHandlerFunc)
	}

	mux.Handle("/", GraphHandlerFunc)

	return &mux, nil
}

func WithObserver(observer *obs.Observer) GraphMuxOption {
	return func(mux *GraphMux) {
		mux.Observer = observer
	}
}

func (mux *GraphMux) Use(mw Middleware) {
	mux.middlewareChain = append(mux.middlewareChain, mw)
}

func (mux *GraphMux) HandleGraphHttpRequest(w http.ResponseWriter, r *http.Request) {
	page, ok := mux.RouteMap[r.URL.Path]
	if ok {
		err := page.Render(w)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.NotFound(w, r)
	}
}

func (mux *GraphMux) SyncGraph(updateChan chan gr.UpdateMessage, errChan chan error) {
	for {
		select {
		case updateMsg, ok := <-updateChan:
			if !ok {
				//TODO: not do
			}
			mux.RouteMap = hr.CreatePathMap(mux.Root)
			switch updateMsg.Type {
			case gr.UpdateTypeCreate:
				mux.logNodeCreation(updateMsg.Webpage)
			case gr.UpdateTypeDelete:
				mux.logNodeDeletion(updateMsg.Webpage)
			}
			// case err, ok <- errChan:

		}
	}
}

func (mux *GraphMux) logNodeCreation(webpage hr.HyperRenderer) {
	if mux.Observer != nil {
		mux.Observer.LogNode(obs.NodeLog{
			ID:        obs.NodeID(webpage.GetID()),
			CreatedAt: time.Now().UTC(),
			DeletedAt: nil,
		})
	}
}

func (mux *GraphMux) logNodeDeletion(webpage hr.HyperRenderer) {
	if mux.Observer != nil {
		now := time.Now().UTC()
		if logNode, ok := mux.Observer.NodeLogMap[obs.NodeID(webpage.GetID())]; ok {
			logNode.DeletedAt = &now
		}
	}
}

// 		if mux.Observer != nil {
// 			mux.Observer.LogVisit(obs.VisitLog{
// 				RemoteAddr: obs.IPAddr(ip),
// 				NodeID:     obs.NodeID(page.GetID()),
// 				VisitedAt:  time.Now().UTC(),
// 			})
// 		}
// 		ip, _, err := net.SplitHostPort(r.RemoteAddr)
// 		if err != nil {
// 			return
// 		}

//
// func (mux *GraphMux) ActivateObserver() {
// 	if mux.Observer == nil {
// 		//TODO: should return error
// 	}
// 	mux.HandleFunc("/observer/age", func(w http.ResponseWriter, r *http.Request) {
// 		query := r.URL.Query()
// 		if ip := query.Get("ip"); ip != "" {
// 			result := fmt.Sprintf("freshness = %v", mux.Observer.GetFreshness(ip, time.Now()))
// 			fmt.Fprintf(w, result)
// 		}
// 	})
// }
