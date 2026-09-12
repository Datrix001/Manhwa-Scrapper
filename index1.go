package main

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"sync"

// 	"github.com/gocolly/colly"
// )

// type ChapterImage struct {
// 	Number int
// 	URL    string
// }

// func main() {

// 	// Variable for Comic Name
// 	var comicName string

// 	// Variable to Get Top Chapter
// 	var topChapterCounter bool

// 	// Stores chapter images in the order they are found
// 	var chapterImages []ChapterImage

// 	c := colly.NewCollector(
// 		colly.AllowedDomains(
// 			"asurascans.com",
// 			"cdn.asurascans.com",
// 		),
// 	)

// 	// --------------------------------------------------
// 	// GET COMIC
// 	// --------------------------------------------------

// 	c.OnHTML(`a[href^="/comics"][href$="53fc8424"]`, func(h *colly.HTMLElement) {

// 		// Only run this on the comics listing page
// 		if h.Request.URL.Path != "/comics/" {
// 			return
// 		}

// 		// Reset top chapter state for each comic
// 		topChapterCounter = false

// 		// Reset chapter images for each comic
// 		chapterImages = nil

// 		link := h.Attr("href")

// 		fmt.Printf("Found Link: %v\n", link)

// 		if err := h.Request.Visit(h.Request.AbsoluteURL(link)); err != nil {
// 			log.Println("failed to visit comic:", err)
// 		}
// 	})

// 	// --------------------------------------------------
// 	// GET TOTAL CHAPTERS
// 	// --------------------------------------------------

// 	c.OnHTML("h2.text-lg.font-bold", func(h *colly.HTMLElement) {

// 		if h.Request.URL.Path == "/comics/" {
// 			return
// 		}

// 		fmt.Printf("Chapter: %s\n", h.Text)
// 	})

// 	// --------------------------------------------------
// 	// GET AUTHOR
// 	// --------------------------------------------------

// 	c.OnHTML(`a[href^="/browse?author"]`, func(h *colly.HTMLElement) {

// 		if h.Request.URL.Path == "/comics/" {
// 			return
// 		}

// 		fmt.Printf("Author Name: %s\n", h.Text)
// 	})

// 	// --------------------------------------------------
// 	// GET COMIC NAME
// 	// --------------------------------------------------

// 	c.OnHTML("article h1", func(h *colly.HTMLElement) {

// 		comicName = strings.TrimSpace(h.Text)

// 		fmt.Println(comicName)
// 	})

// 	// --------------------------------------------------
// 	// GET COVER IMAGE
// 	// --------------------------------------------------

// 	c.OnHTML("#cover-viewer-img", func(h *colly.HTMLElement) {

// 		if h.Request.URL.Path == "/comics/" {
// 			return
// 		}

// 		link := h.Attr("data-full-src")

// 		fmt.Printf("Cover image link: %s\n", link)

// 		if err := h.Request.Visit(link); err != nil {
// 			log.Println("failed to visit cover:", err)
// 		}
// 	})

// 	// --------------------------------------------------
// 	// GET CHAPTER IMAGES
// 	// --------------------------------------------------

// 	c.OnHTML(`div.select-none img.w-full.block`, func(h *colly.HTMLElement) {

// 		if h.Request.URL.Path == "/comics/" {
// 			return
// 		}

// 		link := h.Attr("src")

// 		// Number image according to the order
// 		// in which it appears in the HTML.
// 		number := len(chapterImages) + 1

// 		chapterImages = append(chapterImages, ChapterImage{
// 			Number: number,
// 			URL:    link,
// 		})

// 		fmt.Printf(
// 			"Found chapter image %03d: %s\n",
// 			number,
// 			link,
// 		)
// 	})

// 	// --------------------------------------------------
// 	// GET TOP CHAPTER
// 	// --------------------------------------------------

// 	c.OnHTML(`div.divide-y a[href*="/chapter/"]`, func(h *colly.HTMLElement) {

// 		if topChapterCounter {
// 			return
// 		}

// 		topChapterCounter = true

// 		// Reset image list for this chapter
// 		chapterImages = nil

// 		link := h.Attr("href")

// 		fmt.Printf(
// 			"Top Chapter: %s -> %s\n\n",
// 			strings.TrimSpace(h.Text),
// 			link,
// 		)

// 		if err := h.Request.Visit(h.Request.AbsoluteURL(link)); err != nil {
// 			log.Println("failed to visit chapter:", err)
// 		}
// 	})

// 	// --------------------------------------------------
// 	// AFTER CHAPTER PAGE IS SCRAPED
// 	// --------------------------------------------------

// 	c.OnScraped(func(r *colly.Response) {

// 		// Only run after a chapter page
// 		if !strings.Contains(r.Request.URL.Path, "/chapter/") {
// 			return
// 		}

// 		fmt.Printf(
// 			"\nFound %d chapter images.\n",
// 			len(chapterImages),
// 		)

// 		downloadChapterImages(chapterImages, comicName)
// 	})

// 	// --------------------------------------------------
// 	// HANDLE RESPONSES
// 	// --------------------------------------------------

// 	c.OnResponse(func(r *colly.Response) {

// 		path := r.Request.URL.Path

// 		// Only covers are handled here.
// 		// Chapter images are downloaded separately
// 		// using goroutines.
// 		if strings.Contains(path, "/covers/") {
// 			saveBanner(r, comicName)
// 		}
// 	})

// 	// --------------------------------------------------
// 	// REQUEST LOGGER
// 	// --------------------------------------------------

// 	c.OnRequest(func(r *colly.Request) {

// 		fmt.Println("Visiting", r.URL.String())
// 	})

// 	// --------------------------------------------------
// 	// ERROR HANDLER
// 	// --------------------------------------------------

// 	c.OnError(func(r *colly.Response, err error) {

// 		fmt.Println(
// 			"Request URL:",
// 			r.Request.URL,
// 			"failed with response:",
// 			r,
// 			"\nError:",
// 			err,
// 		)
// 	})

// 	// --------------------------------------------------
// 	// START SCRAPER
// 	// --------------------------------------------------

// 	err := c.Visit("https://asurascans.com/comics/")

// 	if err != nil {
// 		log.Println("Error visiting comics page:", err)
// 	}
// }

// // ======================================================
// // CLEAN COMIC NAME
// // ======================================================

// func cleanName(name string) string {

// 	name = strings.TrimSpace(name)

// 	replacer := strings.NewReplacer(
// 		`\\`, "_",
// 		`/`, "_",
// 		`:`, "_",
// 		`*`, "_",
// 		`?`, "_",
// 		`"`, "_",
// 		`<`, "_",
// 		`>`, "_",
// 		`|`, "_",
// 		" ", "_",
// 	)

// 	return replacer.Replace(name)
// }

// // ======================================================
// // SAVE COVER
// // ======================================================

// func saveBanner(r *colly.Response, comicName string) {

// 	contentType := r.Headers.Get("Content-Type")

// 	if !strings.HasPrefix(contentType, "image/webp") {
// 		return
// 	}

// 	folder := cleanName(comicName)

// 	folder = filepath.Join(
// 		"images",
// 		folder,
// 	)

// 	err := os.MkdirAll(folder, 0755)

// 	if err != nil {
// 		fmt.Println("Error creating folder:", err)
// 		return
// 	}

// 	filename := filepath.Join(
// 		folder,
// 		"cover.webp",
// 	)

// 	err = os.WriteFile(
// 		filename,
// 		r.Body,
// 		0644,
// 	)

// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	fmt.Println("Saved cover:", filename)
// }

// // ======================================================
// // DOWNLOAD ALL CHAPTER IMAGES CONCURRENTLY
// // ======================================================

// func downloadChapterImages(
// 	images []ChapterImage,
// 	comicName string,
// ) {

// 	if len(images) == 0 {
// 		fmt.Println("No chapter images found.")
// 		return
// 	}

// 	folder := cleanName(comicName)

// 	folder = filepath.Join(
// 		"images",
// 		folder,
// 		"Chapters",
// 	)

// 	err := os.MkdirAll(folder, 0755)

// 	if err != nil {
// 		fmt.Println("Error creating folder:", err)
// 		return
// 	}

// 	fmt.Printf(
// 		"Starting %d image downloads...\n\n",
// 		len(images),
// 	)

// 	var wg sync.WaitGroup

// 	// Maximum number of simultaneous downloads
// 	semaphore := make(chan struct{}, 15)

// 	for _, image := range images {

// 		wg.Add(1)

// 		go func(image ChapterImage) {

// 			defer wg.Done()

// 			// Acquire a download slot
// 			semaphore <- struct{}{}

// 			// Release the slot when finished
// 			defer func() {
// 				<-semaphore
// 			}()

// 			downloadChapterImage(
// 				image,
// 				folder,
// 			)

// 		}(image)
// 	}

// 	// Wait for all downloads
// 	wg.Wait()

// 	fmt.Println("\nFinished downloading chapter!")
// }

// // ======================================================
// // DOWNLOAD ONE CHAPTER IMAGE
// // ======================================================

// func downloadChapterImage(
// 	image ChapterImage,
// 	folder string,
// ) {

// 	filename := fmt.Sprintf(
// 		"%03d.webp",
// 		image.Number,
// 	)

// 	path := filepath.Join(
// 		folder,
// 		filename,
// 	)

// 	fmt.Printf(
// 		"Downloading %03d: %s\n",
// 		image.Number,
// 		image.URL,
// 	)

// 	response, err := http.Get(image.URL)

// 	if err != nil {
// 		fmt.Printf(
// 			"Download error %03d: %v\n",
// 			image.Number,
// 			err,
// 		)

// 		return
// 	}

// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {

// 		fmt.Printf(
// 			"HTTP error %03d: %s\n",
// 			image.Number,
// 			response.Status,
// 		)

// 		return
// 	}

// 	data, err := io.ReadAll(response.Body)

// 	if err != nil {

// 		fmt.Printf(
// 			"Read error %03d: %v\n",
// 			image.Number,
// 			err,
// 		)

// 		return
// 	}

// 	err = os.WriteFile(
// 		path,
// 		data,
// 		0644,
// 	)

// 	if err != nil {

// 		fmt.Printf(
// 			"Write error %03d: %v\n",
// 			image.Number,
// 			err,
// 		)

// 		return
// 	}

// 	fmt.Printf(
// 		"Saved %03d.webp\n",
// 		image.Number,
// 	)
// }
