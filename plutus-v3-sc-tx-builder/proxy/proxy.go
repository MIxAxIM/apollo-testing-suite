package proxy

import (
	"net/http"
	"net/url"
)

// GetProxyTransport returns an http.RoundTripper that uses the configured proxy.
func GetProxyTransport(proxyURLStr string) (http.RoundTripper, error) {
	if proxyURLStr == "" {
		return http.DefaultTransport, nil
	}

	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		return nil, err
	}

	return &http.Transport{Proxy: http.ProxyURL(proxyURL)}, nil
}