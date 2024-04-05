package graphmultiplexer

import (
	"fmt"
	"net"
	"net/http"
	"time"

	dsh "github.com/sdqri/sequined/internal/dashboard"
	ggr "github.com/sdqri/sequined/internal/graphgenerator"
	hyr "github.com/sdqri/sequined/internal/hyperrenderer"
	obs "github.com/sdqri/sequined/internal/observer"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type GraphMux struct {
	Root     *hyr.Webpage
	RouteMap map[string]hyr.HyperRenderer
	Observer *obs.Observer

	*http.ServeMux
	middlewareChain  []Middleware
	GraphHandlerFunc http.HandlerFunc
}

type GraphMuxOption func(*GraphMux)

func New(root *hyr.Webpage, opts ...GraphMuxOption) (*GraphMux, error) {
	routeMap := hyr.CreatePathMap(root)

	mux := GraphMux{
		Root:     root,
		RouteMap: routeMap,

		ServeMux:        http.NewServeMux(),
		middlewareChain: make([]Middleware, 0),
	}

	for _, opt := range opts {
		opt(&mux)
	}

	if mux.Observer != nil {
		mux.middlewareChain = append(mux.middlewareChain, VisitLoggerMiddleware(&mux))

		var err error
		hyr.Traverse(mux.Root, func(node hyr.HyperRenderer) bool {
			currentPage, ok := node.(*hyr.Webpage)
			if !ok {
				err = ggr.ErrUnexpectedNodeType
				return true
			}
			mux.logNodeCreation(currentPage)
			return false
		})
		if err != nil {
			return nil, err
		}
	}

	mux.GraphHandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.HandleGraphHttpRequest(w, r)
	})

	for _, mw := range mux.middlewareChain {
		mux.GraphHandlerFunc = mw(mux.GraphHandlerFunc)
	}

	mux.Handle("/", mux.GraphHandlerFunc)

	return &mux, nil
}

func WithObserver(observer *obs.Observer) GraphMuxOption {
	return func(mux *GraphMux) {
		mux.Observer = observer
	}
}

func WithMiddleware(mw Middleware) GraphMuxOption {
	return func(mux *GraphMux) {
		mux.middlewareChain = append(mux.middlewareChain, mw)
	}
}

func (mux *GraphMux) Use(mw Middleware) {
	mux.GraphHandlerFunc = mw(mux.GraphHandlerFunc)
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

func (mux *GraphMux) SyncGraph(updateChan chan ggr.UpdateMessage, errChan chan error) {
	for {
		select {
		case updateMsg, ok := <-updateChan:
			if !ok {
				//TODO: not do
			}
			mux.RouteMap = hyr.CreatePathMap(mux.Root)
			switch updateMsg.Type {
			case ggr.UpdateTypeCreate:
				mux.logNodeCreation(updateMsg.Webpage)
			case ggr.UpdateTypeDelete:
				mux.logNodeDeletion(updateMsg.Webpage)
			}
			// case err, ok <- errChan:

		}
	}
}

func (mux *GraphMux) logNodeCreation(webpage hyr.HyperRenderer) {
	if mux.Observer != nil {
		mux.Observer.LogNode(obs.NodeLog{
			ID:        obs.NodeID(webpage.GetID()),
			CreatedAt: time.Now().UTC(),
			DeletedAt: nil,
		})
	}
}

func (mux *GraphMux) logNodeDeletion(webpage hyr.HyperRenderer) {
	if mux.Observer != nil {
		now := time.Now().UTC()
		if logNode, ok := mux.Observer.NodeLogMap[obs.NodeID(webpage.GetID())]; ok {
			logNode.DeletedAt = &now
		}
	}
}

func (mux *GraphMux) logVisit(req *http.Request) {
	if mux.Observer == nil {
		return
	}

	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	if node, ok := mux.RouteMap[req.URL.Path]; ok {
		if currentPage, ok := node.(*hyr.Webpage); ok {
			mux.Observer.LogVisit(obs.VisitLog{
				RemoteAddr: obs.IPAddr(ip),
				NodeID:     obs.NodeID(currentPage.GetID()),
				VisitedAt:  time.Now().UTC(),
			})
		}
	}
}

func VisitLoggerMiddleware(mux *GraphMux) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next(w, r)

			mux.logVisit(r)
		}
	}
}

func (mux *GraphMux) ActivateDashboard(dashboard *dsh.Dashboard) {
	dashboard.HandleBy(mux.ServeMux)
}
