package checks

import (
	"net/url"
	"strings"
)

// URL may be missing the scheme://
func ParseHostnameFromURL(urlString string) (string, error) {
	urlWithScheme := urlString
	if !strings.Contains(urlString, "://") {
		// missing scheme, add one to allow url.Parse to work correctly
		urlWithScheme = "http://" + urlWithScheme
	}
	parsedUrl, err := url.Parse(urlWithScheme)
	if err != nil {
		return "", err
	}
	return parsedUrl.Hostname(), nil
}
