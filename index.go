package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"
)

var chapterImageCounter int

func main() {
	// Variable for Comic Name
	var comicName string

	// Variable to Get Top Chapter
	var topChapterCounter bool

	// Variable for chapter Image number

	c := colly.NewCollector(
		colly.AllowedDomains("asurascans.com", "cdn.asurascans.com"),

		// colly.MaxDepth(1),
	)

	// c.OnResponse(func(r *colly.Response) {
	// 	fmt.Println("RESPONSE:", r.Request.URL)

	// 	fmt.Println(string(r.Body))
	// })

	//Get Chapter Name
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
		topChapterCounter = false

		link := h.Attr("href")

		fmt.Printf("Found Link: %v\n", link)

		h.Request.Visit(h.Request.AbsoluteURL(link))

	})

	//Get Total Chapter
	c.OnHTML("h2.text-lg.font-bold", func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}

		fmt.Printf("Chapter: %s\n", h.Text)
	})

	//Get Author Name
	c.OnHTML(`a[href^="/browse?author"]`, func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}
		fmt.Printf("Author Name: %s\n", h.Text)
	})

	//Get Comic Name
	c.OnHTML("article h1", func(h *colly.HTMLElement) {
		comicName = h.Text
		println(comicName)
	})

	//Get Cover Image
	c.OnHTML("#cover-viewer-img", func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}

		link := h.Attr("data-full-src")

		fmt.Printf("Image link: %s\n", link)
		h.Request.Visit(link)
	})

	c.OnHTML(`div.select-none img.w-full.block`, func(h *colly.HTMLElement) {
		if h.Request.URL.Path == "/comics/" {
			return
		}
		link := h.Attr("src")
		fmt.Printf("Image link: %s\n", link)
		h.Request.Visit(link)
	})

	//Get Chapters
	c.OnHTML(`div.divide-y a[href*="/chapter/"]`, func(h *colly.HTMLElement) {
		if topChapterCounter {
			return
		}

		topChapterCounter = true
		chapterImageCounter = 0

		link := h.Attr("href")

		fmt.Printf("Top Chapter: %s -> %s\n\n\n\n\n",
			strings.TrimSpace(h.Text),
			link,
		)

		if err := h.Request.Visit(link); err != nil {
			log.Println("failed to visit chapter:", err)
		}
	})

	//Image download
	c.OnResponse(func(r *colly.Response) {
		path := r.Request.URL.Path

		if strings.Contains(path, "/covers/") {
			saveBanner(r, comicName)
			return
		}

		if strings.Contains(path, "/chapters/") {
			saveChapterImage(r, comicName)
			return
		}
	})
	// c.on
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

func saveBanner(r *colly.Response, comicName string) {
	contentType := r.Headers.Get("Content-Type")

	if contentType != "image/webp" {
		return
	}

	folder := cleanName(comicName)
	folder = fmt.Sprintf("images/%s", folder)

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	filename := fmt.Sprintf("%s/cover.webp", folder)

	err = os.WriteFile(filename, r.Body, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Saved image!")
}

func saveChapterImage(r *colly.Response, comicName string) {
	contentType := r.Headers.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/webp") {
		return
	}

	folder := cleanName(comicName)
	folder = fmt.Sprintf("images/%s/Chapters", folder)

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}
	chapterImageCounter++
	filename := fmt.Sprintf("%03d.webp", chapterImageCounter)
	path := filepath.Join(folder, filename)

	err = os.WriteFile(path, r.Body, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Saved chapter image:", path)
}
