package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
)

// Controller
type Controller struct {
	ui      *UI
	channel chan checks.CheckResult

	checks map[string]checks.CheckResult
	state  bool
}

func NewController(ui *UI) *Controller {
	return &Controller{
		channel: make(chan checks.CheckResult, 10),
		ui:      ui,
		checks:  make(map[string]checks.CheckResult),
	}
}

func (c *Controller) GetChan() chan checks.CheckResult {
	return c.channel
}

func (c *Controller) updateState(cr checks.CheckResult) {
	c.checks[cr.Type] = cr
	c.state = true

	for _, res := range c.checks {
		if !res.Success {
			c.state = false
			break
		}
	}
}

func (c *Controller) Init() {
	go func() {
		for {
			r := <-c.channel
			c.updateState(r)

			// When nmtui is shown the UI is suspended, so
			// let's skip any update
			if c.ui.IsNMTuiActive() {
				continue
			}

			// Update the widgets
			switch r.Type {
			case checks.CheckTypeReleaseImagePull:
				c.ui.app.QueueUpdateDraw(func() {
					if r.Success {
						c.ui.markCheckSuccess(0, 0)
					} else {
						c.ui.markCheckFail(0, 0)
						c.ui.appendNewErrorToDetails("Release image pull error", r.Details)
					}
				})
			case checks.CheckTypeReleaseImageHostDNS:
				c.ui.app.QueueUpdateDraw(func() {
					if r.Success {
						c.ui.markCheckSuccess(1, 0)
					} else {
						c.ui.markCheckFail(1, 0)
						c.ui.appendNewErrorToDetails("nslookup failure", r.Details)
					}
				})
			case checks.CheckTypeReleaseImageHostPing:
				c.ui.app.QueueUpdateDraw(func() {
					if r.Success {
						c.ui.markCheckSuccess(2, 0)
					} else {
						c.ui.markCheckFail(2, 0)
						c.ui.appendNewErrorToDetails("ping failure", r.Details)
					}
				})
			case checks.CheckTypeReleaseImageHttp:
				c.ui.app.QueueUpdateDraw(func() {
					if r.Success {
						c.ui.markCheckSuccess(3, 0)
					} else {
						c.ui.markCheckFail(3, 0)
						c.ui.appendNewErrorToDetails("http server not responding", r.Details)
					}
				})
			}

			// A check failed while waiting for the countdown. Timeout dialog must be stopped
			if !c.state && c.ui.isTimeoutDialogActive() {
				c.ui.app.QueueUpdate(func() {
					c.ui.cancelUserPrompt()
				})
			}
		}
	}()
}
