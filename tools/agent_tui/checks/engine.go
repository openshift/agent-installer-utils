package checks

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	CheckTypeReleaseImagePull     = "ReleaseImagePull"
	CheckTypeReleaseImageHostDNS  = "ReleaseImageHostDNS"
	CheckTypeReleaseImageHostPing = "ReleaseImageHostPing"
)

type Config struct {
	ReleaseImageURL  string
	RendezvousHostIP string
}

// ChecksEngine is the model part, and is composed by a number
// of different checks.
// Each Check has a type, frequency and evaluation loop.
// Different checks could have the same type

type CheckResult struct {
	Type    string
	Success bool
	Details string // In case of failure
}

type Check struct {
	Type string
	Freq time.Duration //Note: a ticker could be useful
	Run  func(c chan CheckResult, Freq time.Duration)
}

type ChecksEngine struct {
	checks []*Check
	c      chan CheckResult
}

type checkFunction func() ([]byte, error)

func createCheckResult(f checkFunction, checkType string) CheckResult {
	output, err := f()
	var result CheckResult
	if err != nil {
		result = CheckResult{
			Type:    checkType,
			Success: false,
			Details: string(output),
		}
	} else {
		result = CheckResult{
			Type:    checkType,
			Success: true,
			Details: "",
		}
	}
	return result
}

func newRegistryImagePullCheck(releaseImageURL string) *Check {
	ctype := CheckTypeReleaseImagePull
	return &Check{
		Type: ctype,
		Freq: 5 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				checkFunction := func() ([]byte, error) {
					return exec.Command("podman", "pull", releaseImageURL).CombinedOutput()
				}
				c <- createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func newReleaseImageHostDNSCheck(hostname string) *Check {
	ctype := CheckTypeReleaseImageHostDNS
	return &Check{
		Type: ctype,
		Freq: 5 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				checkFunction := func() ([]byte, error) {
					return exec.Command("nslookup", hostname).CombinedOutput()
				}
				c <- createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func newReleaseImageHostPingCheck(hostname string) *Check {
	ctype := CheckTypeReleaseImageHostPing
	return &Check{
		Type: ctype,
		Freq: 5 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				var checkFunction func() ([]byte, error)
				if hostname == "quay.io" {
					// quay.io does not respond to ping
					checkFunction = func() ([]byte, error) {
						return nil, nil
					}
				} else {
					checkFunction = func() ([]byte, error) {
						return exec.Command("ping", "-c", "4", hostname).CombinedOutput()
					}
				}
				c <- createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func NewChecksEngine(c chan CheckResult, config Config) *ChecksEngine {
	checks := []*Check{}

	hostname, err := ParseHostnameFromURL(config.ReleaseImageURL)
	if err != nil {
		fmt.Printf("Error parsing hostname from releaseImageURL: %s\n", config.ReleaseImageURL)
	}

	checks = append(checks,
		newRegistryImagePullCheck(config.ReleaseImageURL),
		newReleaseImageHostDNSCheck(hostname),
		newReleaseImageHostPingCheck(hostname))

	return &ChecksEngine{
		checks: checks,
		c:      c,
	}
}

func (ce *ChecksEngine) Init() {
	for _, chk := range ce.checks {
		go chk.Run(ce.c, chk.Freq)
	}
}
