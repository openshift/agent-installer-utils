package net

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strings"
	"sync"
)

const (
	USE_CACHE          bool = true
	TO_STDOUT          bool = true
	STDOUT_RED_COLOR        = "\033[1;31m"
	STDOUT_GREEN_COLOR      = "\033[32m"
	STDOUT_BLACK_COLOR      = "\033[0m"
	TCELL_RED_COLOR         = "[red]"
	TCELL_GREEN_COLOR       = "[green]"
	TCELL_BLACK_COLOR       = "[black]"
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

var wg sync.WaitGroup

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

func (v *Validations) HasConnectivityIssue() bool {
	if v.ReleaseImagePullError != "" ||
		v.RendezvousHostPingError != "" ||
		v.ReleaseImageDomainNameResolutionError != "" ||
		v.ReleaseImageHostPingError != "" {
		return true
	} else {
		return false
	}
}

func (v *Validations) checkRendezvousHostConnectivity() string {
	output, err := exec.Command("ping", "-c", "4", v.RendezvousHostIP).CombinedOutput()
	if err != nil {
		v.RendezvousHostPingError = string(output)
	} else {
		v.RendezvousHostPingError = ""
	}
	return v.RendezvousHostPingError
}

func (v *Validations) printRendezvousHostConnectivityStatus(w io.Writer, useCachedValues bool, forStdout bool) {
	if !useCachedValues {
		v.checkRendezvousHostConnectivity()
	}
	if v.RendezvousHostPingError != "" {
		fmt.Fprint(w, v.getRendezvousHostConnectivityFailText(forStdout))
	} else {
		fmt.Fprint(w, v.getRendezvousHostConnectivitySuccessText(forStdout))
	}
}

func (v *Validations) checkReleaseImageConnectivity() string {
	output, err := exec.Command("podman", "pull", v.ReleaseImageURL).CombinedOutput()
	if err != nil {
		v.ReleaseImagePullError = string(output)
	} else {
		v.ReleaseImagePullError = ""
	}
	return v.ReleaseImagePullError
}

func (v *Validations) printReleaseImageConnectivityStatus(w io.Writer, useCachedValues bool, forStdout bool) {
	if !useCachedValues {
		v.checkReleaseImageConnectivity()
	}
	if v.ReleaseImagePullError != "" {
		fmt.Fprint(w, v.getReleaseImageConnectivityFailText(forStdout))
	} else {
		fmt.Fprint(w, v.getReleaseImageConnectivitySuccessText(forStdout))
	}
}

func (v *Validations) checkReleaseImageDomainNameResolution() string {
	output, err := exec.Command("nslookup", v.ReleaseImageDomainName).CombinedOutput()
	if err != nil {
		v.ReleaseImageDomainNameResolutionError = string(output)
	} else {
		v.ReleaseImageDomainNameResolutionError = ""
	}
	return v.ReleaseImageDomainNameResolutionError
}

func (v *Validations) printReleaseImageDomainNameResolutionStatus(w io.Writer, useCachedValues bool, forStdout bool) {
	if !useCachedValues {
		v.checkReleaseImageDomainNameResolution()
	}
	if v.ReleaseImageDomainNameResolutionError != "" {
		fmt.Fprint(w, v.getReleaseImageDomainNameResolutionFailText(forStdout))
	} else {
		fmt.Fprint(w, v.getReleaseImageDomainNameResolutionSuccessText(forStdout))
	}
}

func (v *Validations) checkReleaseImageHostPing() string {
	output, err := exec.Command("ping", "-c", "4", v.ReleaseImageDomainName).CombinedOutput()
	if err != nil {
		v.ReleaseImageHostPingError = string(output)
	} else {
		v.ReleaseImageHostPingError = ""
	}
	return v.ReleaseImageHostPingError
}

func (v *Validations) printReleaseImageHostPingStatus(w io.Writer, useCachedValues bool, forStdout bool) {
	if v.ReleaseImageDomainName == "quay.io" {
		fmt.Fprint(w, "INFO: quay.io does not respond to ping\n")
		v.ReleaseImageHostPingError = ""
		return
	}
	if !useCachedValues {
		v.checkReleaseImageHostPing()
	}
	if v.ReleaseImageHostPingError != "" {
		fmt.Fprint(w, v.getReleaseImageHostPingFailText(forStdout))
	} else {
		fmt.Fprint(w, v.getReleaseImageHostPingSuccessText(forStdout))
	}
}

func getSuccessColor(forStdout bool) string {
	if forStdout {
		return STDOUT_GREEN_COLOR
	} else {
		return TCELL_GREEN_COLOR
	}
}

func getFailColor(forStdout bool) string {
	if forStdout {
		return STDOUT_RED_COLOR
	} else {
		return TCELL_RED_COLOR
	}
}

func getNormalColor(forStdout bool) string {
	if forStdout {
		return STDOUT_BLACK_COLOR
	} else {
		return TCELL_BLACK_COLOR
	}
}

func (v *Validations) getRendezvousHostConnectivityFailText(forStdout bool) string {
	return fmt.Sprintf("ping rendezvous host at %s %sfail%s\n", v.RendezvousHostIP, getFailColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getRendezvousHostConnectivitySuccessText(forStdout bool) string {
	return fmt.Sprintf("ping rendezvous host at %s %ssuccess%s\n", v.RendezvousHostIP, getSuccessColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageConnectivityFailText(forStdout bool) string {
	return fmt.Sprintf("Pull release image %sfail%s\n", getFailColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageConnectivitySuccessText(forStdout bool) string {
	return fmt.Sprintf("Pull release image %ssuccess%s\n", getSuccessColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageDomainNameResolutionFailText(forStdout bool) string {
	return fmt.Sprintf("nslookup release image host at %s %sfail%s\n", v.ReleaseImageDomainName, getFailColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageDomainNameResolutionSuccessText(forStdout bool) string {
	return fmt.Sprintf("nslookup release image host at %s %ssuccess%s\n", v.ReleaseImageDomainName, getSuccessColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageHostPingFailText(forStdout bool) string {
	return fmt.Sprintf("ping release image host at %s %sfail%s\n", v.ReleaseImageDomainName, getFailColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) getReleaseImageHostPingSuccessText(forStdout bool) string {
	return fmt.Sprintf("ping release image host at %s %ssuccess%s\n", v.ReleaseImageDomainName, getSuccessColor(forStdout), getNormalColor(forStdout))
}

func (v *Validations) numberOfErrors() int {
	numberOfErrors := 0
	if v.RendezvousHostPingError != "" {
		numberOfErrors++
	}
	if v.ReleaseImagePullError != "" {
		numberOfErrors++
	}
	if v.ReleaseImageDomainNameResolutionError != "" {
		numberOfErrors++
	}
	if v.ReleaseImageHostPingError != "" {
		numberOfErrors++
	}
	return numberOfErrors
}

func formatError(errorType string, thisErrorNumber int, totalNumberOfErrors int, forStdout bool) string {
	return fmt.Sprintf("%v=== Error %v/%v: %v ===%v\n", getFailColor(forStdout), thisErrorNumber, totalNumberOfErrors, errorType, getNormalColor(forStdout))
}

func (v *Validations) PrintConnectivityStatus(w io.Writer, useCachedValues bool, forStdout bool) {
	goodConnectivity := true
	wg.Add(4)

	if !useCachedValues {
		fmt.Fprintln(w, "Running connectivity checks. Please wait...")
	}

	go func() {
		defer wg.Done()
		v.printRendezvousHostConnectivityStatus(w, useCachedValues, forStdout)
		if v.RendezvousHostPingError != "" {
			goodConnectivity = false
		}
	}()

	go func() {
		defer wg.Done()
		v.printReleaseImageConnectivityStatus(w, useCachedValues, forStdout)
		if v.ReleaseImagePullError != "" {
			goodConnectivity = false
		}
	}()

	go func() {
		defer wg.Done()
		v.printReleaseImageDomainNameResolutionStatus(w, useCachedValues, forStdout)
		if v.ReleaseImageDomainNameResolutionError != "" {
			goodConnectivity = false
		}
	}()

	go func() {
		defer wg.Done()
		v.printReleaseImageHostPingStatus(w, useCachedValues, forStdout)
		if v.ReleaseImageHostPingError != "" {
			goodConnectivity = false
		}
	}()

	wg.Wait()

	if goodConnectivity {
		fmt.Fprintf(w, "%sConnectivity checks successful%s\n", getSuccessColor(forStdout), getNormalColor(forStdout))
	} else {
		numberOfErrors := v.numberOfErrors()
		checksString := "checks"
		if numberOfErrors == 1 {
			checksString = "check"
		}
		fmt.Fprintf(w, "%s%v connectivity %v failed%s\n", getFailColor(forStdout), numberOfErrors, checksString, getNormalColor(forStdout))
		totalNumberOfErrors := v.numberOfErrors()
		errorNumber := 0
		if v.RendezvousHostPingError != "" {
			errorNumber++
			fmt.Fprint(w, formatError("ping rendezvous host error", errorNumber, totalNumberOfErrors, forStdout))
			fmt.Fprintf(w, "%s\n", v.RendezvousHostPingError)
		}
		if v.ReleaseImagePullError != "" {
			errorNumber++
			fmt.Fprint(w, formatError("Pull release image error", errorNumber, totalNumberOfErrors, forStdout))
			fmt.Fprintf(w, "%s\n", v.ReleaseImagePullError)
		}
		if v.ReleaseImageDomainNameResolutionError != "" {
			errorNumber++
			fmt.Fprint(w, formatError("nslookup release image host error", errorNumber, totalNumberOfErrors, forStdout))
			fmt.Fprintf(w, "%s\n", v.ReleaseImageDomainNameResolutionError)
		}
		if v.ReleaseImageHostPingError != "" {
			errorNumber++
			fmt.Fprint(w, formatError("ping release image host error", errorNumber, totalNumberOfErrors, forStdout))
			fmt.Fprintf(w, "%s\n", v.ReleaseImageHostPingError)
		}
	}
}
