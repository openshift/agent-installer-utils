package ui

import (
	"fmt"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP_SAVE_SUCCESS string = "rendezvousIPSaveSuccessPage"
	BACK_BUTTON                            = "<Back>"

	CONNECTIVITY_CHECK_FAIL_TEXT = "Connectivity check failed.\ncurl %s was unsuccessful.\n Check the rendezvous node has booted and shows the login prompt. Or check your network configuration.\n"
	successText                  = "Successfully saved %s as the Rendezvous node IP. "
	otherNodesText               = "Enter %s as the Rendezvous node IP on the other nodes that will form the cluster."
)

func (u *UI) showRendezvousIPSaveSuccessModal(savedIP, connectivityErrorText string, focusForBackButton func()) {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.rendezvousIPSaveSuccessModal = tview.NewModal()
	u.rendezvousIPSaveSuccessModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == OK_BUTTON {
			u.app.Stop()
		}
		if buttonLabel == BACK_BUTTON {
			focusForBackButton()
		}
		if buttonLabel == CONFIGURE_NETWORK_BUTTON {
			u.ShowNMTUI()
		}

	})
	u.rendezvousIPSaveSuccessModal.SetBackgroundColor(newt.ColorGray) // ContrastBackgroundColor, default is newt.ColorBlue
	u.rendezvousIPSaveSuccessModal.SetBorder(true)
	u.rendezvousIPSaveSuccessModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{OK_BUTTON, BACK_BUTTON}
	if connectivityErrorText != "" {
		userPromptButtons = append(userPromptButtons, CONFIGURE_NETWORK_BUTTON)
	}
	u.rendezvousIPSaveSuccessModal.AddButtons(userPromptButtons)

	text := successText
	isRendezvousNode := contains(u.hostIPAddresses(), savedIP)
	if isRendezvousNode {
		// This node was designated as the Rendezvous node.
		// Prompt the user to enter the IP on other nodes.
		text += otherNodesText
	}

	if connectivityErrorText != "" {
		text += "\n\n" + connectivityErrorText
	}

	if isRendezvousNode {
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
