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
	FIELD_ENTER_RENDEZVOUS_IP   = "Rendezvous node IP"
	SAVE_RENDEZVOUS_IP_BUTTON   = "<Save Rendezvous IP>"
	SELECT_IP_ADDRESS_BUTTON    = "<Choose one of this node's IPs to be the Rendezvous node IP>"
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

	rendezvousIPFormDescription := "If you have already designated a node to be the rendezvous node, enter its IP address in the field below."
	rendezvousTextFlex := u.createTextFlex(rendezvousIPFormDescription)
	rendezvousTextNumRows := 5

	u.rendezvousIPForm.AddInputField(FIELD_ENTER_RENDEZVOUS_IP, "", 55, nil, nil)
	u.rendezvousIPForm.SetFieldTextColor(newt.ColorGray)

	list := tview.NewList()
	for i, ip := range u.hostIPAddresses() {
		list.AddItem(ip, "", rune(i), nil)
	}

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
			err := saveRendezvousIPAddress(ipAddress)
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

	selectFormDescription := "If the rendezvous node has not been designated, this node can be designated as the rendezvous node by selecting one of its IP addresses to be the rendezvous node IP."
	selectTextFlex := u.createTextFlex(selectFormDescription)
	selectTextNumRows := 5

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

func (u *UI) createSelectHostIPPage() {
	u.selectIPList = tview.NewList()
	backOption := "<Back>"
	options := append(u.hostIPAddresses(), backOption)
	for _, ip := range options {
		u.selectIPList.AddItem(ip, "", rune(0), func() {
			if ip == backOption {
				u.setFocusToRendezvousIP()
			} else {
				err := saveRendezvousIPAddress(ip)
				if err != nil {
					u.ShowErrorDialog(fmt.Sprintf(saveRendezvousIPError, err.Error()))
				} else {
					u.showRendezvousIPSaveSuccessModal(ip)
				}
			}
		})
	}
	u.selectIPList.SetSelectedBackgroundColor(newt.ColorRed)
	u.selectIPList.SetSelectedTextColor(newt.ColorGray)
	u.selectIPList.ShowSecondaryText(false)
	u.selectIPList.SetBorderPadding(0, 0, 2, 2)

	descriptionText := fmt.Sprintf("Select an IP address from this node to be the Rendezvous node IP.")
	textFlex := u.createTextFlex(descriptionText)
	textRows := 4

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textFlex, textRows, 0, false).
		AddItem(u.selectIPList, len(options)+1, 0, false)
	mainFlex.SetTitle("  Select an IP address to be the Rendezvous node IP  ").
		SetTitleColor(newt.ColorRed).
		SetBorder(true)

	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, len(options)+3+textRows, 0, false).
		AddItem(nil, 0, 1, false)

	width := 80
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, width, 1, false).
		AddItem(nil, 0, 1, false)

	u.pages.AddPage(PAGE_SET_NODE_AS_RENDEZVOUS, flex, true, true)
}

func validateIP(ipAddress string) string {
	if net.ParseIP(ipAddress) == nil {
		return fmt.Sprintf("%s is not a valid IP address", ipAddress)
	}
	return ""
}

func (u *UI) hostIPAddresses() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		u.logger.Errorf("Could not fetch host IPs: %v", err)
	}
	ipv4 := []string{}
	ipv6 := []string{}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && !ipnet.IP.IsLinkLocalUnicast() {
				ipv4 = append(ipv4, ipnet.IP.String())
			} else if ipnet.IP.To16() != nil {
				ipv6 = append(ipv6, ipnet.IP.String())
			}
		}
	}
	u.logger.Infof("current host IPv4 addresses: %v", ipv4)
	u.logger.Infof("current host IPv6 addresses: %v", ipv6)
	return append(ipv4, ipv6...)
}
