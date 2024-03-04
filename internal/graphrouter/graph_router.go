package graphrouter

import (
	"fmt"
	"net/http"

	hr "github.com/sdqri/sequined/internal/hyperrenderer"
)

func Run(addr string, root hr.HyperRenderer) error {
	pathMap := hr.CreateMap(root)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		// Find the page with the given ID
		// You might need to implement a function to search for a page by ID
		fmt.Println(pathMap)
		if page, ok := pathMap[r.URL.Path]; ok {
			err := page.Render(w)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Println("Server started at", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}
