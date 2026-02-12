package ui

import (
	"fmt"
	"net"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	EMPTY_OPTION                        = "" // list option used as spacing between IP addresses and <Back> button
	RENDEZVOUS_CONFIGURE_NETWORK_BUTTON = "<Configure Network>"
)

func (u *UI) createSelectHostIPPage() {
	u.selectIPList = tview.NewList()
	u.refreshSelectIPList()
	u.selectIPList.SetSelectedBackgroundColor(newt.ColorRed)
	u.selectIPList.SetSelectedTextColor(newt.ColorGray)
	u.selectIPList.ShowSecondaryText(false)
	u.selectIPList.SetBorderPadding(0, 0, 2, 2)
	// only use the custom InputCapture if there are IP addresses
	if len(u.hostIPAddresses()) > 0 {
		u.selectIPList.SetInputCapture(getSelectIPListInputCapture(u.selectIPList))
	}

	descriptionText := "Select an IP address from this node to be the rendezvous node IP."
	textFlex := u.createTextFlex(descriptionText)
	textRows := 3

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textFlex, textRows, 0, false).
		AddItem(u.selectIPList, u.selectIPList.GetItemCount()+1, 0, false)
	mainFlex.SetTitle("  Rendezvous node IP selection  ").
		SetTitleColor(newt.ColorRed).
		SetBorder(true)

	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, u.selectIPList.GetItemCount()+3+textRows, 0, false).
		AddItem(nil, 0, 1, false)

	width := 80
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(innerFlex, width, 1, false).
		AddItem(nil, 0, 1, false)

	u.pages.AddPage(PAGE_SET_NODE_AS_RENDEZVOUS, flex, true, true)
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

// This custom InputCapture is used when there are IP addresses.
// When IP addresses are present, a blank line is added between
// the IP addresses and the <Back> and <Configure Network> buttons.
// The InputCapture skips the blank line when navigating the list.
func getSelectIPListInputCapture(list *tview.List) (capture func(event *tcell.EventKey) *tcell.EventKey) {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// List.GetCurrentItem actually returns not the current selected item but the
		// item that was last selected
		previousItemIndex := list.GetCurrentItem()

		switch event.Key() {
		case tcell.KeyTab, tcell.KeyDown, tcell.KeyRight:
			currentItemIndex := previousItemIndex + 1
			if previousItemIndex == list.GetItemCount()-1 {
				// reached the bottom of the selectIPList
				currentItemIndex = 0
			}
			currentItem, _ := list.GetItemText(currentItemIndex)
			if currentItem == EMPTY_OPTION {
				// move the current index up one place to skip the EMPTY_OPTION
				updatedItemIndex := currentItemIndex
				if updatedItemIndex > list.GetItemCount() {
					// reached the bottom of the list
					updatedItemIndex = 0
				}
				list.SetCurrentItem(updatedItemIndex)
			}
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			currentItemIndex := previousItemIndex - 1
			if currentItemIndex < 0 {
				// reached the top of the selectIPList
				currentItemIndex = list.GetItemCount() - 1
			}
			currentItem, _ := list.GetItemText(currentItemIndex)
			if currentItem == EMPTY_OPTION {
				// move the current index down one place to skip the EMPTY_OPTION
				updatedItemIndex := currentItemIndex
				if updatedItemIndex < 0 {
					updatedItemIndex = list.GetItemCount() - 1
				}
				list.SetCurrentItem(updatedItemIndex)
			}
		}

		return event
	}
}

func (u *UI) refreshSelectIPList() {
	u.updateSelectIPList(u.hostIPAddresses())
}

func (u *UI) updateSelectIPList(ipAddresses []string) {
	u.selectIPList.Clear()
	backOption := "<Back>"
	options := ipAddresses
	if len(ipAddresses) > 0 {
		// only add spacer line if there are IP addresses
		options = append(options, EMPTY_OPTION)
	}
	options = append(options, backOption, RENDEZVOUS_CONFIGURE_NETWORK_BUTTON)
	for _, selected := range options {
		u.selectIPList.AddItem(selected, "", rune(0), func() {
			switch selected {
			case EMPTY_OPTION:
				// spacing between IP addresses and buttons
			case backOption:
				u.setFocusToRendezvousIP()
			case RENDEZVOUS_CONFIGURE_NETWORK_BUTTON:
				u.showNMTUIWithErrorDialog(func() {
					u.refreshSelectIPList()
					u.setFocusToSelectIP()
				})
				u.setFocusToSelectIP()
			default:
				err := u.saveRendezvousIPAddress(selected)
				if err != nil {
					u.showRendezvousModal(fmt.Sprintf(SAVE_RENDEZVOUS_IP_ERROR_FORMAT, err.Error()), []string{BACK_BUTTON})
				} else {
					u.showRendezvousIPSaveSuccessModal(selected, u.setFocusToSelectIP)
				}
			}
		})
	}
}
