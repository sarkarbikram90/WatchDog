package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestScrapeWebsite(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve a sample HTML page with a title
		_, _ = w.Write([]byte("<html><head><title>Test Title</title></head><body></body></html>"))
	}))
	defer testServer.Close()

	// Create a wait group and a channel
	var wg sync.WaitGroup
	websiteCh := make(chan Website)

	// Call the scrapeWebsite function with the test server URL
	wg.Add(1)
	go scrapeWebsite(testServer.URL, &wg, websiteCh)

	// Wait for the goroutine to finish
	wg.Wait()

	// Read the result from the channel
	website := <-websiteCh

	// Verify the title was scraped correctly
	if website.Title != "Test Title" {
		t.Errorf("Expected title to be 'Test Title', got '%s'", website.Title)
	}
}
