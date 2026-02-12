package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP_CONNECTIVITY_FAIL string = "rendezvousIPConnectivityFailPage"
	SAVE_AND_CONTINUE_BUTTON             string = "<Save and Continue>"

	CONNECTIVITY_CHECK_FAIL_TEXT_FORMAT string = "Warning: the specified rendezvous IP was not found or yet active."
)

func (u *UI) showRendezvousIPConnectivityFailModal(ipAddress string, focusForBackButton func()) {
	u.connectivityFailModal = tview.NewModal()
	u.connectivityFailModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == SAVE_AND_CONTINUE_BUTTON {
			u.saveRendezvousIPAndShowModalIfError(ipAddress, false)
		}
		if buttonLabel == BACK_BUTTON {
			focusForBackButton()
		}
		if buttonLabel == RENDEZVOUS_CONFIGURE_NETWORK_BUTTON {
			u.showNMTUIWithErrorDialog(u.setFocusToRendezvousIP)
			u.pages.SwitchToPage(PAGE_RENDEZVOUS_IP_CONNECTIVITY_FAIL)
		}

	})
	u.connectivityFailModal.SetBackgroundColor(newt.ColorGray) // ContrastBackgroundColor, default is newt.ColorBlue
	u.connectivityFailModal.SetBorder(true)
	u.connectivityFailModal.SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{SAVE_AND_CONTINUE_BUTTON, BACK_BUTTON, RENDEZVOUS_CONFIGURE_NETWORK_BUTTON}
	u.connectivityFailModal.AddButtons(userPromptButtons)
	u.connectivityFailModal.SetText(CONNECTIVITY_CHECK_FAIL_TEXT_FORMAT)

	u.pages.AddPage(PAGE_RENDEZVOUS_IP_CONNECTIVITY_FAIL, u.connectivityFailModal, true, false)
	u.app.SetFocus(u.connectivityFailModal)
	u.pages.ShowPage(PAGE_RENDEZVOUS_IP_CONNECTIVITY_FAIL)
}
