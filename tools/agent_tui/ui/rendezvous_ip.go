package ui

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP        = "rendezvousIPScreen"
	FIELD_CURRENT_HOST_IP     = "Current node IPs"
	FIELD_SET_IP              = "Use current node IP as rendezvous"
	FIELD_RENDEZVOUS_HOST_IP  = "Rendezvous node IP"
	SAVE_RENDEZVOUS_IP_BUTTON = "<Save and continue>"
	RENDEZVOUS_HOST_ENV_PATH  = "/etc/assisted/rendezvous-host.env"
)

func (u *UI) createRendezvousIPPage(config checks.Config) {
	primaryCheck := tview.NewTable()
	primaryCheck.SetBorder(true)
	primaryCheck.SetTitle("  Current Host IPs  ")
	primaryCheck.SetBorderColor(newt.ColorBlack)
	primaryCheck.SetBackgroundColor(newt.ColorGray)
	primaryCheck.SetTitleColor(newt.ColorBlack)

	u.rendezvousIPForm = tview.NewForm()
	u.rendezvousIPForm.SetBorder(false)
	u.rendezvousIPForm.SetBackgroundColor(newt.ColorGray)
	u.rendezvousIPForm.SetButtonsAlign(tview.AlignCenter)
	u.rendezvousIPForm.AddInputField(FIELD_CURRENT_HOST_IP, strings.Join(u.hostIPAddresses()[:2], ","), 55, nil, nil)
	u.rendezvousIPForm.AddInputField(FIELD_RENDEZVOUS_HOST_IP, "", 55, nil, nil)
	u.rendezvousIPForm.AddCheckbox(FIELD_SET_IP, false, func(checked bool) {
		field := u.rendezvousIPForm.GetFormItemByLabel(FIELD_RENDEZVOUS_HOST_IP).(*tview.InputField)
		if checked && len(u.hostIPAddresses()) > 0 {
			field.SetText(u.hostIPAddresses()[0])
		} else {
			// unchecked, reset rendezvous IP field
			field.SetText("")
		}
	})

	u.rendezvousIPForm.AddButton(SAVE_RENDEZVOUS_IP_BUTTON, func() {
		// save rendezvous IP address and switch to checks page
		ipAddress := u.rendezvousIPForm.GetFormItemByLabel(FIELD_RENDEZVOUS_HOST_IP).(*tview.InputField).GetText()
		validationError := validateIP(ipAddress)
		if validationError != "" {
			u.ShowErrorDialog(fmt.Sprintf(invalidIPText, ipAddress))
		} else {
			err := saveRendezvousIPAddress(ipAddress)
			if err != nil {
				u.ShowErrorDialog(fmt.Sprintf(saveRendezvousIPError, err.Error()))
			} else {
				// set focus to checks page and let controller know rendezvousIP is set
				u.setIsRendezousIPFormActive(false)
				u.setFocusToChecks()
			}
		}
	})
	u.rendezvousIPForm.SetButtonActivatedStyle(tcell.StyleDefault.Background(newt.ColorRed).
		Foreground(newt.ColorGray))
	u.rendezvousIPForm.SetButtonStyle(tcell.StyleDefault.Background(newt.ColorGray).
		Foreground(newt.ColorBlack))

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(u.rendezvousIPForm, 8+2, 0, false)
	mainFlex.SetTitle("  Rendezvous Host IP Setup  ").
		SetTitleColor(newt.ColorRed).
		SetBorder(true).
		SetBackgroundColor(newt.ColorGray).
		SetBorderColor(tcell.ColorBlack)

	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(mainFlex, mainFlexHeight+2, 0, false).
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

	u.pages.SetBackgroundColor(newt.ColorBlue)
	u.pages.AddPage(PAGE_RENDEZVOUS_IP, flex, true, true)
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
			if ipnet.IP.To4() != nil {
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

func saveRendezvousIPAddress(ipAddress string) error {
	_, err := os.Stat(RENDEZVOUS_HOST_ENV_PATH)
	if os.IsNotExist(err) {
		// TODO: Should we not expect RENDEZVOUS_HOST_ENV_PATH to always exist?
		// If so, this block can be removed.
		file, err := os.OpenFile(RENDEZVOUS_HOST_ENV_PATH, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not create and/or write to %s", RENDEZVOUS_HOST_ENV_PATH)
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("NODE_ZERO_IP=%s\nSERVICE_BASE_URL=http://%s:8090/\nIMAGE_SERVICE_BASE_URL=http://%s:8888/\n", ipAddress, ipAddress, ipAddress))
		if err != nil {
			return fmt.Errorf("error writing NODE_ZERO_IP to %s", RENDEZVOUS_HOST_ENV_PATH)
		}
	} else {
		cmd := exec.Command("sed", "-i", fmt.Sprintf("s/^NODE_ZERO_IP=.*/NODE_ZERO_IP=%s/", ipAddress), RENDEZVOUS_HOST_ENV_PATH)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("NODE_ZERO_IP update failed: %v out: %v", err, string(output))
		}

		cmd = exec.Command("sed", "-i", fmt.Sprintf("s/^SERVICE_BASE_URL=.*/SERVICE_BASE_URL=http:\\/\\/%s:8090\\//", ipAddress), RENDEZVOUS_HOST_ENV_PATH)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("SERVICE_BASE_URL update failed: %v out: %v", err, string(output))
		}

		cmd = exec.Command("sed", "-i", fmt.Sprintf("s/^IMAGE_SERVICE_BASE_URL=.*/IMAGE_SERVICE_BASE_URL=http:\\/\\/%s:8888\\//", ipAddress), RENDEZVOUS_HOST_ENV_PATH)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("IMAGE_SERVICE_BASE_URL update failed: %v out: %v", err, string(output))
		}
	}
	return nil
}
