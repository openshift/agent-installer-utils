package ui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	OK_BUTTON         string = "<Ok>"
	PAGE_ERROR_DIALOG        = "invalidIPAddress"

	invalidIPText         = "The IP address %s is not a valid IPv4 or IPv6 address"
	saveRendezvousIPError = "Error saving rendezvous IP address to /etc/assisted/rendezvous-host.env: %v"
)

// Creates the invalid IP address modal but does not add the modal
// to pages. The rendezvousIPForm does that when it validates the IP
// address entered.
func (u *UI) createErrorModal() {
	// view is the modal asking the user if they would still
	// like to change their network configuration.
	u.errorModal = tview.NewModal().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == OK_BUTTON {
				u.setFocusToRendezvousIP()
			}
		}).
		SetBackgroundColor(newt.ColorBlack)
	u.errorModal.
		SetBorderColor(newt.ColorBlack).
		SetBorder(true)
	u.errorModal.
		SetButtonBackgroundColor(newt.ColorGray).
		SetButtonTextColor(newt.ColorRed)
	userPromptButtons := []string{OK_BUTTON}
	u.errorModal.AddButtons(userPromptButtons)

	//u.invalidIPAddressModal.SetText(invalidIPText)
	u.pages.AddPage(PAGE_ERROR_DIALOG, u.errorModal, true, false)
}

func (u *UI) ShowErrorDialog(errorText string) {
	//u.setIsTimeoutDialogActive(true)
	u.errorModal.SetText(errorText)
	u.app.SetFocus(u.errorModal)
	u.pages.ShowPage(PAGE_ERROR_DIALOG)
}
