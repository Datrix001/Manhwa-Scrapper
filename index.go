package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	var comicName string
	c := colly.NewCollector(
		colly.AllowedDomains("asurascans.com", "cdn.asurascans.com"),

		// colly.MaxDepth(1),
	)

	// c.OnResponse(func(r *colly.Response) {
	// 	fmt.Println("RESPONSE:", r.Request.URL)

	// 	fmt.Println(string(r.Body))
	// })

	c.OnHTML(`a[href^="/comics"][href$="53fc8424"]`, func(h *colly.HTMLElement) {
		// link := h.Attr("href")

		// fmt.Printf("Link found: %q -> %s\n", h.Text, link)

		// c.Visit(h.Request.AbsoluteURL(link))

		// if h.Attr("h1 class") == "" {
		// 	return
		// }
		// fmt.Printf("Found it %s \n", h.Text)
		if h.Request.URL.Path != "/comics/" {
			return
		}

		link := h.Attr("href")

		fmt.Printf("Found Link: %v\n", link)

		h.Request.Visit(h.Request.AbsoluteURL(link))

	})

	c.OnHTML("h2.text-lg.font-bold", func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}
		fmt.Printf("Chapter: %s\n", h.Text)
	})

	c.OnHTML(`a[href^="/browse?author"]`, func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}
		fmt.Printf("Author Name: %s\n", h.Text)
	})
	c.OnHTML("article h1", func(h *colly.HTMLElement) {
		comicName = h.Text
		println(comicName)
	})
	c.OnHTML("#cover-viewer-img", func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}

		link := h.Attr("data-full-src")

		fmt.Printf("Image link: %s\n", link)
		h.Request.Visit(link)
	})

	c.OnResponse(func(r *colly.Response) {

		contentType := r.Headers.Get("Content-Type")

		if contentType != "image/webp" {
			return
		}

		folder := cleanName(comicName)
		folder = fmt.Sprintf("images/%s", folder)
		err1 := os.MkdirAll(folder, 0755)
		if err1 != nil {
			fmt.Println("Error creating folder:", err1)
			return
		}
		filename := fmt.Sprintf("%s/cover.webp", folder)

		err := os.WriteFile(filename, r.Body, 0644)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("Saved image!")
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	c.Visit("https://asurascans.com/comics/")

}

func cleanName(name string) string {
	name = strings.TrimSpace(name)

	replacer := strings.NewReplacer(
		`\`, "_",
		`/`, "_",
		`:`, "_",
		`*`, "_",
		`?`, "_",
		`"`, "_",
		`<`, "_",
		`>`, "_",
		`|`, "_",
		" ", "_",
	)

	return replacer.Replace(name)
}
