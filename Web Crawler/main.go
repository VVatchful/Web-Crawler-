package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Define a structure to hold the crawler's settings and state
type Crawler struct {
	client    *http.Client
	visited   map[string]bool
	queue     []string
	maxDepth  int
	currDepth int
	mutex     sync.Mutex
}

// Entry point of the program
func main() {
	// Initialize the HTTP client
	client := &http.Client{}

	// Initialize the crawler's state
	crawler := &Crawler{
		client:   client,
		visited:  make(map[string]bool),
		queue:    []string{"http://example.com"}, // Starting URL
		maxDepth: 3,                              // Set max crawl depth
	}

	// Start the crawling process
	crawler.Start()
}

// Method to start the crawler
func (c *Crawler) Start() {
	for c.currDepth < c.maxDepth {
		// Create a wait group to wait for all goroutines to finish
		var wg sync.WaitGroup
		// Copy the current queue
		currentQueue := c.queue
		c.queue = nil

		// Loop through the queue
		for _, url := range currentQueue {
			// Check if the URL has been visited
			c.mutex.Lock()
			if c.visited[url] {
				c.mutex.Unlock()
				continue
			}
			c.visited[url] = true
			c.mutex.Unlock()

			// Increment the wait group counter
			wg.Add(1)
			// Spawn a new goroutine to handle the URL
			go func(url string) {
				defer wg.Done()
				c.Crawl(url)
			}(url)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Increment the current depth
		c.currDepth++
	}

	fmt.Println("Crawling complete.")
}

// Method to crawl a given URL
func (c *Crawler) Crawl(url string) {
	// Send GET request to the URL
	resp, err := c.client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	// Check if the response status is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK HTTP status: %d\n", resp.StatusCode)
		return
	}

	// Parse the HTML content with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing HTML for %s: %v\n", url, err)
		return
	}

	// Find and process links in the HTML
	c.findLinks(doc, url)
}

// Method to find and process links in the HTML document
func (c *Crawler) findLinks(doc *goquery.Document, base string) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			normalizedURL := c.normalizeURL(link, base)
			if normalizedURL != "" {
				c.mutex.Lock()
				c.queue = append(c.queue, normalizedURL)
				c.mutex.Unlock()
			}
		}
	})
}

// Method to normalize URLs
func (c *Crawler) normalizeURL(href, base string) string {
	u, err := url.Parse(href)
	if err != nil || u.Scheme == "javascript" || u.Scheme == "mailto" {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}
