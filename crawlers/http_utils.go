package crawlers

import (
	"log"
	"net/http"
	"time"
)

func sendGetRequest(url string) (*http.Response, error) {
	return sendRequest(url, "GET")
}

func sendRequest(url, method string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Printf("error when creating request for '%s' with method '%s'\n", url, method)
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:90.0) Gecko/20100101 Firefox/90.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error when querying url %s with method %s: %/s\n", url, method, err)
		return nil, err
	}

	return resp, nil
}
