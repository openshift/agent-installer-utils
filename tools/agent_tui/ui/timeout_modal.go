package ui

import (
	"fmt"
	"time"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	YES_BUTTON string = "<Yes>"
	NO_BUTTON  string = "<No>"
)

// Creates the timeout modal but does not add the modal
// to pages. The activeUserPrompt function does that
// when all checks are successful.
func (u *UI) createTimeoutModal(config checks.Config) {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.timeoutModal = tview.NewModal().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == YES_BUTTON {
				u.cancelUserPrompt()
			} else {
				u.app.Stop()
			}
		}).
		SetBackgroundColor(newt.ColorBlack)
	u.timeoutModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.timeoutModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{YES_BUTTON, NO_BUTTON}
	u.timeoutModal.AddButtons(userPromptButtons)
}

func (u *UI) activateUserPrompt() {
	u.setIsTimeoutDialogActive(true)
	u.app.SetFocus(u.timeoutModal)

	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	timeout := 20 * time.Second

	go func() {
		for {
			select {
			case <-u.timeoutDialogCancel:
				ticker.Stop()
				return

			case t := <-ticker.C:
				elapsed := t.Sub(start)
				if elapsed >= timeout {
					ticker.Stop()
					u.app.Stop()
				}

				modalText := fmt.Sprintf("Agent-based installer connectivity checks passed. No additional network configuration is required. Do you still wish to modify the network configuration for this host?\n\n This prompt will timeout in [red]%.f [white]seconds.", timeout.Seconds()-elapsed.Seconds())
				u.app.QueueUpdateDraw(func() {
					u.timeoutModal.SetText(modalText)
				})
			}
		}
	}()
	u.pages.AddPage("userPromptToConfigureNetworkWith20sTimeout", u.timeoutModal, true, true)
}

func (u *UI) cancelUserPrompt() {
	u.timeoutDialogCancel <- true
	u.setIsTimeoutDialogActive(false)
	u.returnFocusToChecks()
}
