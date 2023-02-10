package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
)

// Controller
type Controller struct {
	ui      *UI
	channel chan checks.CheckResult
	state   State
}

type State struct {
	// default value is false
	ReleaseImagePullSuccess                 bool
	ReleaseImageDomainNameResolutionSuccess bool
	ReleaseImageHostPingSuccess             bool
}

func NewController(ui *UI) *Controller {
	return &Controller{
		channel: make(chan checks.CheckResult, 10),
		ui:      ui,
	}
}

func (c *Controller) GetChan() chan checks.CheckResult {
	return c.channel
}

func (c *Controller) AllChecksSuccess() bool {
	if c.state.ReleaseImagePullSuccess &&
		c.state.ReleaseImageDomainNameResolutionSuccess &&
		c.state.ReleaseImageHostPingSuccess {
		return true
	} else {
		return false
	}
}

func (c *Controller) updateState(cr checks.CheckResult) {
	switch cr.Type {
	case checks.CheckTypeReleaseImagePull:
		c.state.ReleaseImagePullSuccess = cr.Success
	case checks.CheckTypeReleaseImageHostDNS:
		c.state.ReleaseImageDomainNameResolutionSuccess = cr.Success
	case checks.CheckTypeReleaseImageHostPing:
		c.state.ReleaseImageHostPingSuccess = cr.Success
	}
}

func (c *Controller) Init() {
	go func() {
		for {
			r := <-c.channel
			c.updateState(r)

			//Update the widgets
			switch r.Type {
			case checks.CheckTypeReleaseImagePull:
				c.ui.app.QueueUpdate(func() {
					if r.Success {
						c.ui.markCheckSuccess(0, 0)
					} else {
						c.ui.markCheckFail(0, 0)
						c.ui.appendNewErrorToDetails("Release image pull error", r.Details)
					}
				})
			case checks.CheckTypeReleaseImageHostDNS:
				c.ui.app.QueueUpdate(func() {
					if r.Success {
						c.ui.markCheckSuccess(1, 0)
					} else {
						c.ui.markCheckFail(1, 0)
						c.ui.appendNewErrorToDetails("nslookup failure", r.Details)
					}
				})
			case checks.CheckTypeReleaseImageHostPing:
				c.ui.app.QueueUpdate(func() {
					if r.Success {
						c.ui.markCheckSuccess(2, 0)
					} else {
						c.ui.markCheckFail(2, 0)
						c.ui.appendNewErrorToDetails("ping failure", r.Details)
					}
				})
			}

			if c.AllChecksSuccess() {
				c.ui.app.QueueUpdate(func() {
					if !c.ui.activatedUserPrompt {
						// Only activate user prompt once
						c.ui.activateUserPrompt()
						c.ui.activatedUserPrompt = true
					}
				})
			}
			c.ui.app.QueueUpdateDraw(func() {})
		}
	}()
}
