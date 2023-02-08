package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHostnameFromURL(t *testing.T) {
	cases := []struct {
		name             string
		urlToTest        string
		expectedError    error
		expectedHostname string
	}{
		{
			name:             "normal url",
			urlToTest:        "http://www.hostname.com/something/somewhere",
			expectedError:    nil,
			expectedHostname: "www.hostname.com",
		},
		{
			name:             "normal url with port",
			urlToTest:        "http://www.hostname.com:8080/something/somewhere",
			expectedError:    nil,
			expectedHostname: "www.hostname.com",
		},
		{
			name:             "url without scheme",
			urlToTest:        "quay.io/something/somewhere",
			expectedError:    nil,
			expectedHostname: "quay.io",
		},
		{
			name:             "url without scheme with port",
			urlToTest:        "quay.io:8080/something/somewhere",
			expectedError:    nil,
			expectedHostname: "quay.io",
		},
	}

	for _, tc := range cases {
		hostname, err := ParseHostnameFromURL(tc.urlToTest)
		assert.Equal(t, tc.expectedError, err)
		assert.Equal(t, tc.expectedHostname, hostname)
	}
}
