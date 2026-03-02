package ui

import (
	"fmt"
	"time"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	YES_BUTTON                 string = "<Yes>"
	NO_BUTTON                  string = "<No>"
	MODIFY_BUTTON              string = "<Modify>"
	QUIT_TIMEOUT_BUTTON        string = "<Quit>"
	PAGE_TIMEOUTSCREEN         string = "timeout"
	PAGE_RENDEZVOUS_IP_TIMEOUT string = "rendezvousIPTimeout"

	timeout = 20 * time.Second

	modalText = "Agent-based installer connectivity checks passed. No additional network configuration is required." +
		"Do you still wish to modify the network configuration for this host?\n\n " +
		"This prompt will timeout in [red]%.f [white]seconds."
	rendezvousIPTimeoutModalText = "Rendezvous IP has been set to [red]%s[white].\n\n" +
		"Press <Modify> to change it, or <Quit> to exit.\n\n" +
		"This prompt will automatically exit in [red]%.f[white] seconds."
)

// ============================================================================
// Network Configuration Timeout Modal
// ============================================================================

// Creates the timeout modal but does not add the modal
// to pages. The activeUserPrompt function does that
// when all checks are successful.
func (u *UI) createTimeoutModal() {
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
		SetBackgroundColor(newt.ColorGray)
	u.timeoutModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.timeoutModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{YES_BUTTON, NO_BUTTON}
	u.timeoutModal.AddButtons(userPromptButtons)

	u.timeoutModal.SetText(fmt.Sprintf(modalText, timeout.Seconds()))
	u.pages.AddPage(PAGE_TIMEOUTSCREEN, u.timeoutModal, true, false)
}

func (u *UI) ShowTimeoutDialog() {
	u.setIsTimeoutDialogActive(true)
	u.app.SetFocus(u.timeoutModal)
	u.pages.ShowPage(PAGE_TIMEOUTSCREEN)

	// Start countdown timer
	u.startCountdownTimer(timeout, u.timeoutDialogCancel, func(remaining float64) {
		// Update message with remaining time
		u.app.QueueUpdateDraw(func() {
			u.timeoutModal.SetText(fmt.Sprintf(modalText, remaining))
		})
	}, func() {
		// On timeout - exit application
		u.app.Stop()
	})
}

func (u *UI) cancelUserPrompt() {
	u.timeoutDialogCancel <- true
	u.setIsTimeoutDialogActive(false)
	u.setFocusToChecks()
}

// ============================================================================
// Rendezvous IP Timeout Modal
// ============================================================================

// createRendezvousIPTimeoutModal creates the rendezvous IP timeout modal
func (u *UI) createRendezvousIPTimeoutModal() {
	u.rendezvousIPTimeoutModal = tview.NewModal().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				// Modify button - close modal and show form with prefilled IP
				u.cancelRendezvousIPTimeout()
				u.setFocusToRendezvousIP()
			} else {
				// Quit button - exit the application
				u.cancelRendezvousIPTimeout()
				u.app.Stop()
			}
		}).
		SetBackgroundColor(newt.ColorGray)
	u.rendezvousIPTimeoutModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.rendezvousIPTimeoutModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)

	userPromptButtons := []string{MODIFY_BUTTON, QUIT_TIMEOUT_BUTTON}
	u.rendezvousIPTimeoutModal.AddButtons(userPromptButtons)
	u.pages.AddPage(PAGE_RENDEZVOUS_IP_TIMEOUT, u.rendezvousIPTimeoutModal, true, false)
}

func (u *UI) ShowRendezvousIPTimeoutDialog(rendezvousIP string) {
	u.setIsRendezvousIPTimeoutActive(true)
	u.rendezvousIPTimeoutModal.SetText(fmt.Sprintf(rendezvousIPTimeoutModalText, rendezvousIP, timeout.Seconds()))
	u.app.SetFocus(u.rendezvousIPTimeoutModal)
	u.pages.ShowPage(PAGE_RENDEZVOUS_IP_TIMEOUT)

	// Start countdown timer
	u.startCountdownTimer(timeout, u.rendezvousIPTimeoutCancel, func(remaining float64) {
		// Update message with remaining time
		u.app.QueueUpdateDraw(func() {
			u.rendezvousIPTimeoutModal.SetText(fmt.Sprintf(rendezvousIPTimeoutModalText, rendezvousIP, remaining))
		})
	}, func() {
		// On timeout - quit the application
		u.app.QueueUpdateDraw(func() {
			u.setIsRendezvousIPTimeoutActive(false)
			u.pages.HidePage(PAGE_RENDEZVOUS_IP_TIMEOUT)
			u.logger.Infof("Rendezvous IP timeout expired, exiting application")
			u.app.Stop()
		})
	})
}

func (u *UI) cancelRendezvousIPTimeout() {
	u.rendezvousIPTimeoutCancel <- true
	u.setIsRendezvousIPTimeoutActive(false)
	u.pages.HidePage(PAGE_RENDEZVOUS_IP_TIMEOUT)
}

// ============================================================================
// Common Helper Functions
// ============================================================================

// startCountdownTimer runs a countdown timer in a goroutine
// duration: how long until timeout
// cancelChan: channel to cancel the timer
// onTick: called every second with remaining time in seconds
// onTimeout: called when timer expires
func (u *UI) startCountdownTimer(
	duration time.Duration,
	cancelChan chan bool,
	onTick func(remaining float64),
	onTimeout func(),
) {
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-cancelChan:
				return

			case t := <-ticker.C:
				elapsed := t.Sub(start)
				if elapsed >= duration {
					onTimeout()
					return
				}

				remaining := duration.Seconds() - elapsed.Seconds()
				onTick(remaining)
			}
		}
	}()

}
