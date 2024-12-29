package scraper

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
)

// randomly chooses a proxy for the http request
func getRandomProxyTransport() *http.Transport {
	proxySelector := NewRandomProxySelector()
	return &http.Transport{
		Proxy: proxySelector.selectRandomProxy,
	}
}

func NewRandomProxySelector() *RandomProxySelector {
	proxies, err := fetchProxies()
	if err != nil {
		return nil
	}
	return &RandomProxySelector{
		Proxies: proxies,
	}
}

func (t *RandomProxySelector) selectRandomProxy(req *http.Request) (*url.URL, error) {
	if len(t.Proxies) == 0 {
		return nil, fmt.Errorf("no proxies found")
	}
	proxy := t.Proxies[rand.IntN(len(t.Proxies))]
	log.Printf("using proxy: %s for request to %s\n", proxy, req.URL)
	return proxy, nil
}

func fetchProxies() ([]*url.URL, error) {
	proxiesResp, err := http.Get(PROXIES_API_URL)
	if err != nil {
		return nil, fmt.Errorf("fetch proxies error: %v", err)
	}

	proxiesBody, err := io.ReadAll(proxiesResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read proxies response body error: %v", err)
	}
	defer proxiesResp.Body.Close()

	proxiesBodyStr := string(proxiesBody)
	proxiesBodyStr = strings.TrimSuffix(proxiesBodyStr, "\n")

	if len(proxiesBodyStr) == 0 {
		return nil, fmt.Errorf("found no proxies")
	}

	proxies := strings.Split(strings.ReplaceAll(proxiesBodyStr, "\r", "\n"), "\n")
	var availableProxies []*url.URL
	for i := range proxies {
		if proxies[i] != "" {
			urlString := "http://" + proxies[i]
			url, err := url.Parse(urlString)
			if err != nil {
				continue
			}
			availableProxies = append(availableProxies, url) // they should all be http proxies
		}
	}

	return availableProxies, nil
}
