package ui

import (
	"fmt"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP_SAVE_SUCCESS string = "rendezvousIPSaveSuccessPage"
	CONTINUE_BUTTON                 string = "<Continue>"
	BACK_BUTTON                     string = "<Back>"

	SUCCESS_TEXT_FORMAT     = "Successfully saved %s as the rendezvous node IP. "
	OTHER_NODES_TEXT_FORMAT = "Enter %s as the rendezvous node IP on the other nodes that will form the cluster."
)

func (u *UI) showRendezvousIPSaveSuccessModal(savedIP string, focusForBackButton func()) {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.rendezvousIPSaveSuccessModal = tview.NewModal()
	u.rendezvousIPSaveSuccessModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == CONTINUE_BUTTON {
			u.app.Stop()
		}
		if buttonLabel == BACK_BUTTON {
			focusForBackButton()
		}
		if buttonLabel == CONFIGURE_NETWORK_BUTTON {
			u.showNMTUIWithErrorDialog(u.setFocusToRendezvousIP)
			u.pages.SwitchToPage(PAGE_RENDEZVOUS_IP_SAVE_SUCCESS)
		}

	})
	u.rendezvousIPSaveSuccessModal.SetBackgroundColor(newt.ColorGray) // ContrastBackgroundColor, default is newt.ColorBlue
	u.rendezvousIPSaveSuccessModal.SetBorder(true)
	u.rendezvousIPSaveSuccessModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{CONTINUE_BUTTON, BACK_BUTTON}
	u.rendezvousIPSaveSuccessModal.AddButtons(userPromptButtons)

	text := SUCCESS_TEXT_FORMAT
	isRendezvousNode := contains(u.hostIPAddresses(), savedIP)
	if isRendezvousNode {
		// This node was designated as the Rendezvous node.
		// Prompt the user to enter the IP on other nodes.
		text += OTHER_NODES_TEXT_FORMAT
		u.rendezvousIPSaveSuccessModal.SetText(fmt.Sprintf(text, savedIP, savedIP))
	} else {
		u.rendezvousIPSaveSuccessModal.SetText(fmt.Sprintf(text, savedIP))
	}

	u.pages.AddPage(PAGE_RENDEZVOUS_IP_SAVE_SUCCESS, u.rendezvousIPSaveSuccessModal, true, false)
	u.app.SetFocus(u.rendezvousIPSaveSuccessModal)
	u.pages.ShowPage(PAGE_RENDEZVOUS_IP_SAVE_SUCCESS)
}

func contains(arr []string, str string) bool {
	for _, element := range arr {
		if element == str {
			return true
		}
	}
	return false
}
