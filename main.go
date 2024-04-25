package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// Website represents a website to scrape
type Website struct {
	URL       string
	Title     string
	IPAddress string
}

// scrapeWebsite fetches the title and IP address of a website
func scrapeWebsite(url string, wg *sync.WaitGroup, websiteCh chan<- Website) {
	defer wg.Done()

	// Fetch title
	title, err := scrapeTitle(url)
	if err != nil {
		fmt.Println("Error fetching title for", url, ":", err)
		title = "N/A"
	}

	// Fetch IP address
	ipAddr, err := fetchIPAddress(url)
	if err != nil {
		fmt.Println("Error fetching IP address for", url, ":", err)
		ipAddr = "N/A"
	}

	// Send the website information to the channel
	websiteCh <- Website{URL: url, Title: title, IPAddress: ipAddr}
}

// scrapeTitle fetches the title of a website
func scrapeTitle(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	var title string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return title, nil
}

// fetchIPAddress fetches the IP address of a website
func fetchIPAddress(url string) (string, error) {
	host := extractHost(url)
	ipAddr, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	return ipAddr[0].String(), nil
}

// extractHost extracts the host from a URL
func extractHost(url string) string {
	parts := strings.Split(url, "//")
	if len(parts) > 1 {
		return parts[1]
	}
	return parts[0]
}

func main() {
	var urls []string

	fmt.Println("Enter URLs (one per line), type 'done' when finished:")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		var url string
		if !scanner.Scan() {
			break
		}
		url = scanner.Text()
		if url == "done" {
			break
		}
		// Check if the URL has a scheme, if not, prepend "https://"
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + "www." + url
		}
		urls = append(urls, url)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	var wg sync.WaitGroup
	websiteCh := make(chan Website)

	// Start a goroutine for each URL to scrape the website information concurrently
	for _, url := range urls {
		wg.Add(1)
		go scrapeWebsite(url, &wg, websiteCh)
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(websiteCh)
	}()

	// Collect website information from the channel
	for website := range websiteCh {
		fmt.Printf("URL: %s, Title: %s, IP Address: %s\n", website.URL, website.Title, website.IPAddress)
	}
}
