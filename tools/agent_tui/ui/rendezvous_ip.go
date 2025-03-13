package ui

import (
	"fmt"
	"net"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP          = "rendezvousIPScreen"
	PAGE_SET_NODE_AS_RENDEZVOUS = "setNodeAsRendezvousScreen"
	FIELD_ENTER_RENDEZVOUS_IP   = "Rendezvous IP"
	SAVE_RENDEZVOUS_IP_BUTTON   = "<Save Rendezvous IP>"
	SELECT_IP_ADDRESS_BUTTON    = "<Designate this node as the Rendezvous node by selecting one of its IPs>"
	RENDEZVOUS_HOST_ENV_PATH    = "/etc/assisted/rendezvous-host.env"
)

func (u *UI) createTextFlex(text string) *tview.Flex {
	textView := tview.NewTextView()
	textView.SetText(text)
	textView.SetWordWrap(true)

	flex := tview.NewFlex().
		AddItem(textView, 0, 1, false)
	flex.SetBorder(true)
	flex.SetBorderColor(newt.ColorGray) // to keep the spacing but have the border invisible
	return flex
}

func (u *UI) createRendezvousIPPage(config checks.Config) {
	u.rendezvousIPForm = tview.NewForm()
	u.rendezvousIPForm.SetBorder(false)
	u.rendezvousIPForm.SetButtonsAlign(tview.AlignCenter)

	rendezvousIPFormDescription := "Enter the Rendezvous node's IP address if one has been designated."
	rendezvousTextFlex := u.createTextFlex(rendezvousIPFormDescription)
	rendezvousTextNumRows := 3

	u.rendezvousIPForm.AddInputField(FIELD_ENTER_RENDEZVOUS_IP, "", 55, nil, nil)
	u.rendezvousIPForm.SetFieldTextColor(newt.ColorGray)

	u.rendezvousIPForm.AddButton(SAVE_RENDEZVOUS_IP_BUTTON, func() {
		// save rendezvous IP address and switch to checks page
		ipAddress := u.rendezvousIPForm.GetFormItemByLabel(FIELD_ENTER_RENDEZVOUS_IP).(*tview.InputField).GetText()
		validationError := validateIP(ipAddress)
		if validationError != "" {
			if ipAddress == "" {
				ipAddress = "<blank>"
			}
			u.ShowErrorDialog(fmt.Sprintf(invalidIPText, ipAddress))
		} else {
			err := u.saveRendezvousIPAddress(ipAddress)
			if err != nil {
				u.ShowErrorDialog(fmt.Sprintf(saveRendezvousIPError, err.Error()))
			} else {
				u.showRendezvousIPSaveSuccessModal(ipAddress)
			}
		}
	})
	u.rendezvousIPForm.SetButtonActivatedStyle(tcell.StyleDefault.Background(newt.ColorRed).
		Foreground(newt.ColorGray))
	u.rendezvousIPForm.SetButtonStyle(tcell.StyleDefault.Background(newt.ColorGray).
		Foreground(newt.ColorBlack))

	selectFormDescription := "----------------------------------- or ------------------------------------\n\n"
	selectTextFlex := u.createTextFlex(selectFormDescription)
	selectTextNumRows := 3

	u.selectIPForm = tview.NewForm()
	u.selectIPForm.SetBorder(false)
	u.selectIPForm.SetButtonsAlign(tview.AlignCenter)
	u.selectIPForm.AddButton(SELECT_IP_ADDRESS_BUTTON, func() {
		// switch to select IP address page
		u.setFocusToSelectIP()
	})
	u.selectIPForm.SetButtonActivatedStyle(tcell.StyleDefault.Background(newt.ColorRed).
		Foreground(newt.ColorGray))
	u.selectIPForm.SetButtonStyle(tcell.StyleDefault.Background(newt.ColorGray).
		Foreground(newt.ColorBlack))

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(rendezvousTextFlex, rendezvousTextNumRows, 0, false).
		AddItem(u.rendezvousIPForm, 5, 0, false).
		AddItem(selectTextFlex, selectTextNumRows, 0, false).
		AddItem(u.selectIPForm, 4, 0, false)
	mainFlex.SetTitle("  Rendezvous Node IP Setup  ").
		SetTitleColor(newt.ColorRed).
		SetBorder(true)

	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, mainFlexHeight+1+rendezvousTextNumRows+selectTextNumRows, 0, false).
		AddItem(nil, 0, 1, false)

	// Allow the user to cycle the focus only over the configured items
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyUp:
			u.focusedItem++
			if u.focusedItem > len(u.focusableItems)-1 {
				u.focusedItem = 0
			}

		case tcell.KeyBacktab, tcell.KeyDown:
			u.focusedItem--
			if u.focusedItem < 0 {
				u.focusedItem = len(u.focusableItems) - 1
			}

		default:
			// forward the event to the default handler
			return event
		}

		u.app.SetFocus(u.focusableItems[u.focusedItem])
		return nil
	})

	width := 80
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, width, 1, false).
		AddItem(nil, 0, 1, false)

	u.pages.AddPage(PAGE_RENDEZVOUS_IP, flex, true, true)
}

func validateIP(ipAddress string) string {
	if net.ParseIP(ipAddress) == nil {
		return fmt.Sprintf("%s is not a valid IP address", ipAddress)
	}
	return ""
}
