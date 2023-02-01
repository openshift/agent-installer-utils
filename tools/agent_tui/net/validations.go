package net

import (
	"net/url"
	"os/exec"
	"strings"
)

type Validations struct {
	RendezvousHostIP                      string
	RendezvousHostPingError               string
	ReleaseImageURL                       string
	ReleaseImagePullError                 string
	ReleaseImageDomainName                string
	ReleaseImageDomainNameResolutionError string
	ReleaseImageHostPingError             string
}

// URL may be missing the scheme://
func parseHostnameFromURL(urlString string) (string, error) {
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

func NewValidations(releaseImageURL string, rendezvousHostIP string) (*Validations, error) {
	releaseImageHostName, err := parseHostnameFromURL(releaseImageURL)
	if err != nil {
		return nil, err
	}

	return &Validations{
		ReleaseImageURL:        releaseImageURL,
		RendezvousHostIP:       rendezvousHostIP,
		ReleaseImageDomainName: releaseImageHostName,
	}, nil
}

func (v *Validations) CheckConnectivity() {
	output, err := v.checkReleaseImageConnectivity()
	if err != nil {
		v.ReleaseImagePullError = string(output)

		output, err = v.checkReleaseImageDomainNameResolution()
		if err != nil {
			v.ReleaseImageDomainNameResolutionError = string(output)
		} else {
			v.ReleaseImageDomainNameResolutionError = ""
		}

		output, err = v.checkReleaseImageHostPing()
		if err != nil {
			v.ReleaseImageHostPingError = string(output)
		} else {
			v.ReleaseImageHostPingError = ""
		}
	} else {
		v.ReleaseImagePullError = ""
	}

	output, err = v.checkRendezvousHostConnectivity()
	if err != nil {
		v.RendezvousHostPingError = string(output)
	} else {
		v.RendezvousHostPingError = ""
	}
}

func (v *Validations) HasConnectivityIssue() bool {
	if v.ReleaseImagePullError != "" ||
		v.RendezvousHostPingError != "" ||
		v.ReleaseImageDomainNameResolutionError != "" {
		return true
	} else {
		return false
	}
}

func (v *Validations) checkRendezvousHostConnectivity() ([]byte, error) {
	return exec.Command("ping", "-c", "4", v.RendezvousHostIP).CombinedOutput()
}

func (v *Validations) checkReleaseImageConnectivity() ([]byte, error) {
	return exec.Command("podman", "pull", v.ReleaseImageURL).CombinedOutput()
}

func (v *Validations) checkReleaseImageDomainNameResolution() ([]byte, error) {
	return exec.Command("nslookup", v.ReleaseImageDomainName).CombinedOutput()
}

func (v *Validations) checkReleaseImageHostPing() ([]byte, error) {
	return exec.Command("ping", "-c", "4", v.ReleaseImageDomainName).CombinedOutput()
}
