package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type ChapterImage struct {
	Number int
	URL    string
}

type Comic struct {
	Name          string
	Author        string
	Chapter       string
	CoverURL      string
	ChapterURL    string
	ChapterImages []ChapterImage
}

// Stores information about every image that we send
// to the async image collector.
type ImageInfo struct {
	ComicName string
	Page      int
	IsCover   bool
}

var imageInfo = make(map[string]ImageInfo)
var imageInfoMutex sync.RWMutex

func main() {

	var comic Comic

	// ==================================================
	// MAIN COLLECTOR
	// ==================================================

	c := colly.NewCollector(
		colly.AllowedDomains(
			"asurascans.com",
			"cdn.asurascans.com",
		),
	)

	// ==================================================
	// IMAGE COLLECTOR
	// ==================================================

	imageCollector := colly.NewCollector(
		colly.AllowedDomains(
			"cdn.asurascans.com",
		),
		colly.Async(true),
	)

	// Maximum 10 concurrent image downloads
	imageCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10,
	})

	// ==================================================
	// TOP CHAPTER
	// ==================================================

	var topChapterCounter bool

	// ==================================================
	// GET COMIC LINK
	// ==================================================

	c.OnHTML(`a[href^="/comics"][href$="53fc8424"]`, func(h *colly.HTMLElement) {

		if h.Request.URL.Path != "/comics/" {
			return
		}

		topChapterCounter = false

		link := h.Attr("href")

		fmt.Printf(
			"Found Link: %s\n",
			link,
		)

		err := h.Request.Visit(
			h.Request.AbsoluteURL(link),
		)

		if err != nil {
			log.Println(
				"Failed to visit comic:",
				err,
			)
		}
	})

	// ==================================================
	// GET CHAPTER NAME
	// ==================================================

	c.OnHTML("h2.text-lg.font-bold", func(h *colly.HTMLElement) {

		if h.Request.URL.Path == "/comics/" {
			return
		}

		comic.Chapter = strings.TrimSpace(h.Text)

		fmt.Printf(
			"Chapter: %s\n",
			comic.Chapter,
		)
	})

	// ==================================================
	// GET AUTHOR
	// ==================================================

	c.OnHTML(`a[href^="/browse?author"]`, func(h *colly.HTMLElement) {

		if h.Request.URL.Path == "/comics/" {
			return
		}

		comic.Author = strings.TrimSpace(h.Text)

		fmt.Printf(
			"Author Name: %s\n",
			comic.Author,
		)
	})

	// ==================================================
	// GET COMIC NAME
	// ==================================================

	c.OnHTML("article h1", func(h *colly.HTMLElement) {
		comic.ChapterImages = nil
		comic.Name = strings.TrimSpace(h.Text)

		fmt.Printf(
			"Comic Name: %s\n",
			comic.Name,
		)
	})

	// ==================================================
	// GET COVER IMAGE
	// ==================================================

	c.OnHTML("#cover-viewer-img", func(h *colly.HTMLElement) {

		if h.Request.URL.Path == "/comics/" {
			return
		}

		link := h.Attr("data-full-src")

		comic.CoverURL = link

		fmt.Printf(
			"Image link: %s\n",
			link,
		)

		// ----------------------------------------------
		// IMPORTANT
		// Save the comic name with this request.
		// ----------------------------------------------

		storeImageInfo(
			link,
			ImageInfo{
				ComicName: comic.Name,
				IsCover:   true,
			},
		)

		err := imageCollector.Visit(link)

		if err != nil {
			log.Println(
				"Failed to visit cover:",
				err,
			)
		}
	})

	// ==================================================
	// GET CHAPTER IMAGES
	// ==================================================

	c.OnHTML(`div.select-none img.w-full.block`, func(h *colly.HTMLElement) {

		if h.Request.URL.Path == "/comics/" {
			return
		}

		link := h.Attr("src")

		// ----------------------------------------------
		// Assign page number BEFORE downloading
		// ----------------------------------------------

		pageNumber := len(comic.ChapterImages) + 1

		image := ChapterImage{
			Number: pageNumber,
			URL:    link,
		}

		comic.ChapterImages = append(
			comic.ChapterImages,
			image,
		)

		fmt.Printf(
			"Found Page %d: %s\n",
			pageNumber,
			link,
		)

		// ----------------------------------------------
		// Store information for this specific request
		// ----------------------------------------------

		storeImageInfo(
			link,
			ImageInfo{
				ComicName: comic.Name,
				Page:      pageNumber,
				IsCover:   false,
			},
		)

		// ----------------------------------------------
		// Send to ASYNC image collector
		// ----------------------------------------------

		err := imageCollector.Visit(link)

		if err != nil {
			log.Println(
				"Failed to visit image:",
				err,
			)
		}
	})

	// ==================================================
	// GET TOP CHAPTER URL
	// ==================================================

	c.OnHTML(`div.divide-y a[href*="/chapter/"]`, func(h *colly.HTMLElement) {

		if topChapterCounter {
			return
		}

		topChapterCounter = true

		link := h.Attr("href")

		fmt.Printf(
			"Top Chapter: %s -> %s\n\n",
			strings.TrimSpace(h.Text),
			link,
		)

		comic.ChapterURL = link

		err := h.Request.Visit(
			h.Request.AbsoluteURL(link),
		)

		if err != nil {
			log.Println(
				"Failed to visit chapter:",
				err,
			)
		}
	})

	// ==================================================
	// IMAGE COLLECTOR
	// ==================================================

	imageCollector.OnResponse(func(r *colly.Response) {

		url := r.Request.URL.String()

		info, exists := getImageInfo(url)

		if !exists {
			log.Println(
				"Image info not found:",
				url,
			)
			return
		}

		// ----------------------------------------------
		// COVER
		// ----------------------------------------------

		if info.IsCover {

			saveBanner(
				r,
				info.ComicName,
			)

			return
		}

		// ----------------------------------------------
		// CHAPTER IMAGE
		// ----------------------------------------------

		saveChapterImage(
			r,
			info.ComicName,
			info.Page,
		)
	})

	// ==================================================
	// START
	// ==================================================

	err := c.Visit(
		"https://asurascans.com/comics/",
	)

	if err != nil {
		log.Fatal(err)
	}

	// Wait for the main scraper
	c.Wait()

	// Wait for ALL image downloads
	imageCollector.Wait()

	fmt.Printf(
		"\n\nComic:\n%+v\n",
		comic,
	)
}

// ======================================================
// STORE IMAGE INFORMATION
// ======================================================

func storeImageInfo(url string, info ImageInfo) {

	imageInfoMutex.Lock()
	defer imageInfoMutex.Unlock()

	imageInfo[url] = info
}

// ======================================================
// GET IMAGE INFORMATION
// ======================================================

func getImageInfo(url string) (ImageInfo, bool) {

	imageInfoMutex.RLock()
	defer imageInfoMutex.RUnlock()

	info, exists := imageInfo[url]

	return info, exists
}

// ======================================================
// CLEAN NAME
// ======================================================

func cleanName(name string) string {

	name = strings.TrimSpace(name)

	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)

	return replacer.Replace(name)
}

// ======================================================
// SAVE COVER
// ======================================================

func saveBanner(
	r *colly.Response,
	comicName string,
) {

	contentType := r.Headers.Get("Content-Type")

	if !strings.HasPrefix(
		contentType,
		"image/webp",
	) {
		return
	}

	folder := cleanName(comicName)

	folder = filepath.Join(
		"images",
		folder,
	)

	err := os.MkdirAll(
		folder,
		0755,
	)

	if err != nil {
		fmt.Println(
			"Error creating folder:",
			err,
		)
		return
	}

	filename := filepath.Join(
		folder,
		"cover.webp",
	)

	err = os.WriteFile(
		filename,
		r.Body,
		0644,
	)

	if err != nil {
		fmt.Println(
			"Error:",
			err,
		)
		return
	}

	fmt.Println(
		"Saved cover:",
		filename,
	)
}

// ======================================================
// SAVE CHAPTER IMAGE
// ======================================================

func saveChapterImage(
	r *colly.Response,
	comicName string,
	imageNumber int,
) {

	contentType := r.Headers.Get("Content-Type")

	if !strings.HasPrefix(
		contentType,
		"image/webp",
	) {
		return
	}

	folder := cleanName(comicName)

	folder = filepath.Join(
		"images",
		folder,
		"Chapters",
	)

	err := os.MkdirAll(
		folder,
		0755,
	)

	if err != nil {
		fmt.Println(
			"Error creating folder:",
			err,
		)
		return
	}

	filename := fmt.Sprintf(
		"%03d.webp",
		imageNumber,
	)

	path := filepath.Join(
		folder,
		filename,
	)

	err = os.WriteFile(
		path,
		r.Body,
		0644,
	)

	if err != nil {
		fmt.Println(
			"Error:",
			err,
		)
		return
	}

	fmt.Println(
		"Saved chapter image:",
		path,
	)
}
