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
	ReleaseImageHttp                        bool
	lastCheckAllSuccess                     bool
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
	case checks.CheckTypeReleaseImageHttp:
		c.state.ReleaseImageHttp = cr.Success
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
			case checks.CheckTypeReleaseImageHttp:
				c.ui.app.QueueUpdate(func() {
					if r.Success {
						c.ui.markCheckSuccess(3, 0)
					} else {
						c.ui.markCheckFail(3, 0)
						c.ui.appendNewErrorToDetails("http server not responding", r.Details)
					}
				})
			}

			if c.ui.nmtuiActive {
				continue
			}

			allChecksSuccessful := c.AllChecksSuccess()
			if !allChecksSuccessful && c.ui.timeoutDialogActive {
				c.ui.app.QueueUpdate(func() {
					c.ui.cancelUserPrompt()
				})
			}
			if allChecksSuccessful && !c.ui.timeoutDialogActive && c.state.lastCheckAllSuccess != allChecksSuccessful {
				c.ui.app.QueueUpdate(func() {
					c.ui.activateUserPrompt()
				})
			}
			c.ui.app.QueueUpdateDraw(func() {})
			c.state.lastCheckAllSuccess = allChecksSuccessful
		}
	}()
}
