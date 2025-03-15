package ui

import (
	"fmt"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP_SAVE_SUCCESS string = "rendezvousIPSaveSuccessPage"
	CONTINUE_INSTALLATION_BUTTION          = "<Continue with installation>"
	BACK_BUTTON                            = "<Back>"

	successText    = "Successfully saved %s as the Rendezvous node IP. "
	otherNodesText = "Enter %s as the Rendezvous node IP on the other nodes that will form the cluster."
)

func (u *UI) showRendezvousIPSaveSuccessModal(savedIP string, focusForBackButton func()) {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.rendezvousIPSaveSuccessModal = tview.NewModal()
	u.rendezvousIPSaveSuccessModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == CONTINUE_INSTALLATION_BUTTION {
			u.app.Stop()
		}
		if buttonLabel == BACK_BUTTON {
			focusForBackButton()
		}
	})
	u.rendezvousIPSaveSuccessModal.SetBackgroundColor(newt.ColorGray) // ContrastBackgroundColor, default is newt.ColorBlue
	u.rendezvousIPSaveSuccessModal.SetBorder(true)
	u.rendezvousIPSaveSuccessModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{CONTINUE_INSTALLATION_BUTTION, BACK_BUTTON}
	u.rendezvousIPSaveSuccessModal.AddButtons(userPromptButtons)

	text := successText
	if contains(u.hostIPAddresses(), savedIP) {
		// This node was designated as the Rendezvous node.
		// Prompt the user to enter the IP on other nodes.
		text += otherNodesText
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
