package ui

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	PAGE_RENDEZVOUS_IP          = "rendezvousIPScreen"
	PAGE_SET_NODE_AS_RENDEZVOUS = "setNodeAsRendezvousScreen"
	FIELD_ENTER_RENDEZVOUS_IP   = "Rendezvous IP"
	SAVE_RENDEZVOUS_IP_BUTTON   = "<Save rendezvous IP>"
	SELECT_IP_ADDRESS_BUTTON    = "<This is the rendezvous node>"
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

	rendezvousIPFormDescription := "The rendezvous node will be the one managing your cluster installation and where you'll be able to configure all cluster settings.\n\n\n\nI've already obtained the rendezvous IP from another node."
	rendezvousTextFlex := u.createTextFlex(rendezvousIPFormDescription)
	rendezvousTextNumRows := 8

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
			u.showRendezvousModal(fmt.Sprintf(INVALID_IP_TEXT_FORMAT, ipAddress), []string{BACK_BUTTON})
			return
		}

		u.showRendezvousModal(fmt.Sprintf(CHECKING_CONNECTIVITY_TEXT_FORMAT, ipAddress), []string{})

		// run the connectivity check in a goroutine in the background
		// because the "Checking connectivity" modal is displayed only
		// after this function returns.
		go u.displayModalAfterConnectivityCheck(ipAddress)
	})
	u.rendezvousIPForm.SetButtonActivatedStyle(tcell.StyleDefault.Background(newt.ColorRed).
		Foreground(newt.ColorGray))
	u.rendezvousIPForm.SetButtonStyle(tcell.StyleDefault.Background(newt.ColorGray).
		Foreground(newt.ColorBlack))

	orDivider := "                                    or                                    \n\n"
	selectTextFlex := u.createTextFlex(orDivider)
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
	mainFlex.SetTitle("  Rendezvous node setup  ").
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

func (u *UI) checkConnectivity(ipAddress string) bool {
	url := fmt.Sprintf("http://%s:8090/api/assisted-install/v2", ipAddress)
	connectivtyFailedText := ""
	stdout, connectivityErr := exec.Command("curl", "-m 1", url).CombinedOutput()
	if connectivityErr != nil {
		connectivtyFailedText = CONNECTIVITY_CHECK_FAIL_TEXT_FORMAT
		u.logger.Infof("Connectivity check failed: %s: %s", connectivtyFailedText, stdout)
		return false
	}

	u.logger.Infof("has connectivity to %s", ipAddress)
	return true
}

func (u *UI) displayModalAfterConnectivityCheck(ipAddress string) {
	haveConnectivity := u.checkConnectivity(ipAddress)
	u.app.QueueUpdateDraw(func() {
		if !haveConnectivity {
			u.showRendezvousIPConnectivityFailModal(ipAddress, u.setFocusToRendezvousIP)
		} else {
			u.saveRendezvousIPAndShowModalIfError(ipAddress, true)
		}
	})
}

func (u *UI) saveRendezvousIPAndShowModalIfError(ipAddress string, confirmSave bool) {
	err := u.saveRendezvousIPAddress(ipAddress)

	if err != nil {
		u.showRendezvousModal(fmt.Sprintf(SAVE_RENDEZVOUS_IP_ERROR_FORMAT, err.Error()), []string{BACK_BUTTON})
		return
	}

	if confirmSave {
		u.showRendezvousIPSaveSuccessModal(ipAddress, u.setFocusToRendezvousIP)
		return
	}

	u.app.Stop()
}
