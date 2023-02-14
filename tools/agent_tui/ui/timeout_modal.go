package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
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
		SetTextColor(tcell.ColorBlack).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == YES_BUTTON {
				u.cancelUserPrompt()
			} else {
				u.app.Stop()
			}
		}).
		SetBackgroundColor(newt.ColorGray)
	u.timeoutModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.timeoutModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{YES_BUTTON, NO_BUTTON}
	u.timeoutModal.AddButtons(userPromptButtons)
}

func (u *UI) activateUserPrompt(ctx context.Context) {
	u.setIsTimeoutDialogActive(true)
	u.app.SetFocus(u.timeoutModal)
	go func(ctx context.Context) {
		timeoutSeconds := 20
		i := 0
		for i <= timeoutSeconds {
			select {
			case <-ctx.Done():
				return
			default:
				modalText := fmt.Sprint("Agent-based installer connectivity checks passed. No additional network configuration is required. Do you still wish to modify the network configuration for this host?\n\n This prompt will timeout in [red]", timeoutSeconds-i, " [black]seconds.")
				u.app.QueueUpdateDraw(func() {
					u.timeoutModal.SetText(modalText)
				})
				time.Sleep(1 * time.Second)
				i++
			}
		}
	}(ctx)
	u.pages.AddPage("userPromptToConfigureNetworkWith20sTimeout", u.timeoutModal, true, true)
}

func (u *UI) cancelUserPrompt(cancel context.CancelFunc) {
	cancel()
	u.setIsTimeoutDialogActive(false)
	u.returnFocusToChecks()
}
