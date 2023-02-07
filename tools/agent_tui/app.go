package agent_tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/dialogs"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/forms"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/net"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	YES string = "Yes"
	NO  string = "No"
)

func activateNetworkConfigurationScreen(app *tview.Application, pages *tview.Pages) {
	regNodeForm := forms.RegNodeModalForm(app, pages)
	pages.AddPage("regNodeConfig", regNodeForm, true, true)
}

func userPromptHandler(app *tview.Application, pages *tview.Pages, exitAfterTimeout bool) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonLabel == YES {
			exitAfterTimeout = false
			activateNetworkConfigurationScreen(app, pages)
		} else {
			app.Stop()
		}
	}
}

func updateTimeoutText(app *tview.Application, view *tview.Modal, timeout int, exitAfterTimeout bool) {
	i := 0
	for i <= timeout {
		// TODO: fix [black]. It doesn't match the grey text preceeding it.
		modalText := fmt.Sprint("Agent-based installer connectivity checks passed. No additional network configuration is required. Do you still wish to modify the network configuration for this host?\n\n This prompt will timeout in [blue::b]", timeout-i, " [black]seconds.")
		app.QueueUpdateDraw(func() {
			view.SetText(modalText)
		})
		time.Sleep(1 * time.Second)
		i++
	}

	if exitAfterTimeout {
		app.Stop()
	}
}

func App(app *tview.Application) {
	if app == nil {
		app = tview.NewApplication()
	}
	pages := tview.NewPages()

	exitAfterTimeout := false

	background := tview.NewBox().
		SetBorder(false).
		SetBackgroundColor(newt.ColorBlue)

	_, err := net.ValidateConnectivity()

	if err == nil {
		// Connectivity checks passed. Give 20 seconds for user
		// to start network configuration. If there is no input
		// application exits.
		exitAfterTimeout = true

		// view is the modal asking the user if they would still
		// like to change their network configuration.
		view := tview.NewModal().
			SetTextColor(tcell.ColorBlack).
			SetDoneFunc(userPromptHandler(app, pages, exitAfterTimeout)).
			SetBackgroundColor(newt.ColorGray).
			SetButtonTextColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorDarkGray)

		userPromptButtons := []string{YES, NO}
		view.AddButtons(userPromptButtons)

		go updateTimeoutText(app, view, 20, exitAfterTimeout)

		pages.AddPage("background", background, true, true).
			AddPage("userPromptToConfigureNetworkWith20sTimeout", view, true, true)

	} else {
		// Connectivity checks failed. Go directly to the
		// network configuration screen.
		activateNetworkConfigurationScreen(app, pages)
	}

	if err := app.SetRoot(pages, true).Run(); err != nil {
		dialogs.PanicDialog(app, err)
	}
}
