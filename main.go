package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func readRoutes() {
	// Open the routes.txt file and read the routes line by line, and store them in a slice
	file, err := os.Open("routes.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// The routes are stored in this slice... is a datatype
	var routes []string
	for scanner.Scan() {
		// Sort the slice according to the request method in the routes.txt file??
		routes = append(routes, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

// Bearer Run the requests through these proxies: proxies = {'http': 'http://127.0.0.1:8080', 'https': 'http://127.0.0.1:8080'}
// Set up a context with a timeout.
// use the JWT Bearer token for authentication.
// You can get it from BurpSuite repeater. or history tab.
// Place yours in the variable Bearer between the ""
// var Authorization
var Bearer = ""

func fetch() {
	// Create a new http.Client with a timeout of 5 seconds
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create a new http.Request with the given URL and method
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Add the JWT Bearer token to the Authorization header of the request
	req.Header.Add("Authorization", Bearer)

	// Make the request to the given URL and store the response in a variable
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
}

var routes []string

func concurrentFetch() {
	for line := 1; line < len(routes); line++ {
		if line == 0 {
			continue // Skip the first line, which contains the base URL
		}
		start := time.Now()
		ch := make(chan string)
		for _, url := range os.Args[1:] {
			// go fetch and include the JWT Bearer token in the Authorization header for each request
			// and pass each request through the proxies map
			proxies := "http': 'http://127.0.0.1:8080', 'https': 'http://127.0.0.1:8080"
			var transport = &http.Transport{
				Proxy: func(_ *http.Request) (*url.URL, error) { // <- this is problematic, causing errors
					proxyURL, err := url.Parse(proxies["http"])
					/* TODO: Fix this
					url.URL is not a type
					url.Parse undefined (type string has no field or method Parse)
					cannot convert "http" (untyped string constant) to type int
					*/
					if err != nil {
						return nil, err
					}
					return proxyURL, nil
				},
				OnProxyConnectResponse: nil,
				DialContext:            nil,
				Dial:                   nil,
				DialTLSContext:         nil,
				DialTLS:                nil,
				TLSClientConfig:        nil,
				TLSHandshakeTimeout:    0,
				DisableKeepAlives:      false,
				DisableCompression:     false,
				MaxIdleConns:           0,
				MaxIdleConnsPerHost:    0,
				MaxConnsPerHost:        0,
				IdleConnTimeout:        0,
				ResponseHeaderTimeout:  0,
				ExpectContinueTimeout:  0,
				TLSNextProto:           nil,
				ProxyConnectHeader:     nil,
				GetProxyConnectHeader:  nil,
				MaxResponseHeaderBytes: 0,
				WriteBufferSize:        0,
				ReadBufferSize:         0,
				ForceAttemptHTTP2:      false,
			}
			client := &http.Client{Transport: transport}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				ch <- fmt.Sprintf("fetch: %v", err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+Bearer)
			resp, err := client.Do(req)
			if err != nil {
				ch <- fmt.Sprintf("fetch: %v", err)
				continue
			}

			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- fmt.Sprintf("read: %v", err)
				continue
			}
			req, err = http.NewRequest("GET", url, nil)
			if err != nil {
				ch <- fmt.Sprintf("fetch: %v", err)
				continue
			}

			req.Header.Set("Authorization", "Bearer "+Bearer)
			client = &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				ch <- fmt.Sprintf("fetch: %v", err)
				continue
			}
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- fmt.Sprintf("read: %v", err)
				continue
			}
			ch <- fmt.Sprintf("%.2fs  %7d  %s", time.Since(start).Seconds(), resp.ContentLength, url)
			go fetch() // start a goroutine
		}
		for range os.Args[1:] {
			fmt.Println(<-ch) // receive from channel ch
		}
		fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
	}

}

func fetchTwo(url string, ch chan<- string) {
	// resp, err := http.Get(url) and include the Bearer token in the Authorization header for each request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- fmt.Sprintf("fetch: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+Bearer)
	client := &http.Client{}
	resp, err := client.Do(req) // send request to server and get response
	if err != nil {
		ch <- fmt.Sprintf("fetch: %v", err)
		return
	}
	defer resp.Body.Close() // close the response body when we're done
	if resp.StatusCode != http.StatusOK {
		ch <- fmt.Sprintf("fetch: %s %s: %s",
			resp.Status, url, resp.Header.Get("Content-Type"))
		return
	}
	_, err = io.Copy(ioutil.Discard, resp.Body) //
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

}

// The main fuction serves as the entry point and calls all the other functions
func main() {
	// call the read routes.txt function
	readRoutes()
	// call the fetch fuction/ make the requests
	fetch()
	// call the concurrent fetch function
	concurrentFetch()
	// call the fetchTwo function
	fetchTwo("", nil)
}
