package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
)

// Controller
type Controller struct {
	ui      *UI
	channel chan checks.CheckResult
	status  bool
}

func NewController(ui *UI) *Controller {
	return &Controller{
		channel: make(chan checks.CheckResult),
		ui:      ui,
	}
}

func (c *Controller) GetChan() chan checks.CheckResult {
	return c.channel
}

func (c *Controller) Init() {
	go func() {
		for {
			r := <-c.channel

			// TODO: Check if all checks are passing, if yes
			// switch UI to prompt user if they would like to
			// still contine with configuration. exiting automatically
			// after 20 seconds if no response.
			c.status = r.Success
			//...

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
			c.ui.app.Draw()
		}
	}()
}
