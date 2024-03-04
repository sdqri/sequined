package main

import (
	"fmt"

	gr "github.com/sdqri/sequined/internal/graphrouter"
	hr "github.com/sdqri/sequined/internal/hyperrenderer"
)

func main() {
	root := hr.NewWebpage(hr.WebpageTypeHub)
	page2 := hr.NewWebpage(hr.WebpageTypeAuthority)
	page3 := hr.NewWebpage(hr.WebpageTypeAuthority)
	page4 := hr.NewWebpage(hr.WebpageTypeAuthority)
	root.AddLink(page2)
	root.AddLink(page3)
	page2.AddLink(page4)

	page2.Faker().Paragraph(5, 50, 500, "|")

	// Run the HTTP server
	if err := gr.Run(":8080", root); err != nil {
		fmt.Println("Error:", err)
	}
}
