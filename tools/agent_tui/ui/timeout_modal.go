package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	YES string = "Yes"
	NO  string = "No"
)

// Creates the timeout modal but does not add the modal
// to pages. The activeUserPrompt function does that
// when all checks are successful.
func (u *UI) createTimeoutModal(config checks.Config) {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.timeoutModal = tview.NewModal().
		SetTextColor(tcell.ColorBlack).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == YES {
				u.exitAfterTimeout = false
				u.returnFocusToChecks()
			} else {
				u.app.Stop()
			}
		}).
		SetBackgroundColor(newt.ColorGray).
		SetButtonTextColor(tcell.ColorBlack).
		SetButtonBackgroundColor(tcell.ColorDarkGray)

	userPromptButtons := []string{YES, NO}
	u.timeoutModal.AddButtons(userPromptButtons)
}

func (u *UI) activateUserPrompt() {
	u.exitAfterTimeout = true
	u.app.SetFocus(u.timeoutModal)
	go func() {
		timeoutSeconds := 20
		i := 0
		for i <= timeoutSeconds {
			modalText := fmt.Sprint("Agent-based installer connectivity checks passed. No additional network configuration is required. Do you still wish to modify the network configuration for this host?\n\n This prompt will timeout in [blue]", timeoutSeconds-i, " [black]seconds.")
			u.app.QueueUpdateDraw(func() {
				u.timeoutModal.SetText(modalText)
			})
			time.Sleep(1 * time.Second)
			i++
		}

		if u.exitAfterTimeout {
			u.app.Stop()
		}
	}()
	u.pages.AddPage("userPromptToConfigureNetworkWith20sTimeout", u.timeoutModal, true, true)
}
