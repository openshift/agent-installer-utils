package ui

import (
	"fmt"
	"net"

	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

func (u *UI) createSelectHostIPPage() {
	u.selectIPList = tview.NewList()
	u.refreshSelectIPList()
	u.selectIPList.SetSelectedBackgroundColor(newt.ColorRed)
	u.selectIPList.SetSelectedTextColor(newt.ColorGray)
	u.selectIPList.ShowSecondaryText(false)
	u.selectIPList.SetBorderPadding(0, 0, 2, 2)

	descriptionText := fmt.Sprintf("Select an IP address from this node to be the Rendezvous node IP.")
	textFlex := u.createTextFlex(descriptionText)
	textRows := 3

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textFlex, textRows, 0, false).
		AddItem(u.selectIPList, u.selectIPList.GetItemCount()+1, 0, false)
	mainFlex.SetTitle("  Rendezvous Node IP Selection  ").
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

func (u *UI) refreshSelectIPList() {
	listSize := u.selectIPList.GetItemCount()
	for i := 0; i < listSize; i++ {
		u.selectIPList.RemoveItem(i)
	}
	backOption := "<Back>"
	configureNetworkOption := "<Configure Network>"
	options := append(u.hostIPAddresses(), backOption, configureNetworkOption)
	for _, selected := range options {
		u.selectIPList.AddItem(selected, "", rune(0), func() {
			if selected == backOption {
				u.setFocusToRendezvousIP()
			} else if selected == configureNetworkOption {
				u.showNMTUIWithErrorDialog(func() {
					u.refreshSelectIPList()
					u.setFocusToSelectIP()
				})
			} else {
				err := u.saveRendezvousIPAddress(selected)
				if err != nil {
					u.ShowErrorDialog(fmt.Sprintf(saveRendezvousIPError, err.Error()))
				} else {
					u.showRendezvousIPSaveSuccessModal(selected, u.setFocusToSelectIP)
				}
			}
		})
	}
}
