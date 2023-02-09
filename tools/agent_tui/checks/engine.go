package checks

import (
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const (
	CheckTypeReleaseImagePull     = "ReleaseImagePull"
	CheckTypeReleaseImageHostDNS  = "ReleaseImageHostDNS"
	CheckTypeReleaseImageHostPing = "ReleaseImageHostPing"
	CheckTypeAllChecksSuccess     = "AllChecksSuccess"
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

type Engine struct {
	checks  []*Check
	channel chan CheckResult
	state   *State
}

type checkFunction func() ([]byte, error)

type State struct {
	// default value is false
	// RendezvousHostPingSuccess               bool
	ReleaseImagePullSuccess                 bool
	ReleaseImageDomainNameResolutionSuccess bool
	ReleaseImageHostPingSuccess             bool
}

func (e *Engine) AllChecksSucess() bool {
	if e.state.ReleaseImagePullSuccess &&
		// e.state.RendezvousHostPingSuccess &&
		e.state.ReleaseImageDomainNameResolutionSuccess &&
		e.state.ReleaseImageHostPingSuccess {
		return true
	} else {
		return false
	}
}

func (e *Engine) updateState(cr CheckResult) {
	switch cr.Type {
	case CheckTypeReleaseImagePull:
		e.state.ReleaseImagePullSuccess = cr.Success
	case CheckTypeReleaseImageHostDNS:
		e.state.ReleaseImageDomainNameResolutionSuccess = cr.Success
	case CheckTypeReleaseImageHostPing:
		e.state.ReleaseImageHostPingSuccess = cr.Success
	}
}

func (e *Engine) createCheckResult(f checkFunction, checkType string) CheckResult {
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
			Details: string(output),
		}
	}
	e.updateState(result)
	return result
}

func (e *Engine) newAllSuccessCheck() *Check {
	ctype := CheckTypeAllChecksSuccess
	return &Check{
		Type: ctype,
		Freq: 3 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				checkFunction := func() ([]byte, error) {
					if !e.AllChecksSucess() {
						errorString := "not all checks are successful"
						return []byte(errorString), errors.New(errorString)
					} else {
						return nil, nil
					}
				}
				c <- e.createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func (e *Engine) newRegistryImagePullCheck(releaseImageURL string) *Check {
	ctype := CheckTypeReleaseImagePull
	return &Check{
		Type: ctype,
		Freq: 5 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				checkFunction := func() ([]byte, error) {
					return exec.Command("podman", "pull", releaseImageURL).CombinedOutput()
				}
				c <- e.createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func (e *Engine) newReleaseImageHostDNSCheck(hostname string) *Check {
	ctype := CheckTypeReleaseImageHostDNS
	return &Check{
		Type: ctype,
		Freq: 5 * time.Second,
		Run: func(c chan CheckResult, freq time.Duration) {
			for {
				checkFunction := func() ([]byte, error) {
					return exec.Command("nslookup", hostname).CombinedOutput()
				}
				c <- e.createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func (e *Engine) newReleaseImageHostPingCheck(hostname string) *Check {
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
				c <- e.createCheckResult(checkFunction, ctype)
				time.Sleep(freq)
			}
		},
	}
}

func NewEngine(c chan CheckResult, config Config) *Engine {
	state := &State{}
	checks := []*Check{}

	hostname, err := ParseHostnameFromURL(config.ReleaseImageURL)
	if err != nil {
		fmt.Printf("Error parsing hostname from releaseImageURL: %s\n", config.ReleaseImageURL)
	}

	e := &Engine{
		checks:  checks,
		channel: c,
		state:   state,
	}

	e.checks = append(e.checks,
		e.newRegistryImagePullCheck(config.ReleaseImageURL),
		e.newReleaseImageHostDNSCheck(hostname),
		e.newReleaseImageHostPingCheck(hostname),
		e.newAllSuccessCheck())

	return e
}

func (e *Engine) Init() {
	for _, chk := range e.checks {
		go chk.Run(e.channel, chk.Freq)
	}
}
