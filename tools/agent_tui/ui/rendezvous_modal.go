package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	OK_BUTTON             = "<Ok>"
	PAGE_RENDEZVOUS_MODAL = "rendezvousModal"

	invalidIPText             = "The IP address %s is not a valid IPv4 or IPv6 address"
	checkingConnectivityText  = "Checking connectivity to %s"

	saveRendezvousIPError     = "Error saving rendezvous IP address to /etc/assisted/rendezvous-host.env: %v"
)

// A generic modal used for Rendezvous IP confirmations
// Buttons are added to the modal depending on context using
// the ShowRendezvousModal(text, buttons) function.
func (u *UI) createRendezvousModal() {
	u.rendezvousModal = tview.NewModal().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case OK_BUTTON:
				u.setFocusToRendezvousIP()
			case CONFIGURE_NETWORK_BUTTON:
				u.ShowNMTUI()
			}
		}).
		SetBackgroundColor(newt.ColorGray)
	u.rendezvousModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.rendezvousModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)

	u.pages.AddPage(PAGE_RENDEZVOUS_MODAL, u.rendezvousModal, true, false)
}

func (u *UI) ShowRendezvousModal(text string, buttons []string) {
	u.rendezvousModal.ClearButtons()
	u.rendezvousModal.AddButtons(buttons)
	u.rendezvousModal.SetText(text)
	u.app.SetFocus(u.rendezvousModal)
	u.pages.ShowPage(PAGE_RENDEZVOUS_MODAL)
}
