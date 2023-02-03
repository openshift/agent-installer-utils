package agent_tui

import (
	"fmt"
	"os"
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

func activateNetworkConfigurationScreen(app *tview.Application, pages *tview.Pages, validations *net.Validations) {
	regNodeForm := forms.RegNodeModalForm(app, pages, validations)
	pages.AddPage("regNodeConfig", regNodeForm, true, true)
}

func userPromptHandler(app *tview.Application, pages *tview.Pages, validations *net.Validations, exitAfterTimeout bool) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonLabel == YES {
			exitAfterTimeout = false
			activateNetworkConfigurationScreen(app, pages, validations)
		} else {
			app.Stop()
		}
	}
}

func updateTimeoutText(app *tview.Application, view *tview.Modal, timeout int, exitAfterTimeout bool) {
	i := 0
	for i <= timeout {
		modalText := fmt.Sprint("Agent-based installer connectivity checks passed. No additional network configuration is required. Do you still wish to modify the network configuration for this host?\n\n This prompt will timeout in [blue]", timeout-i, " [black]seconds.")
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

	validations, err := net.NewValidations(os.Getenv("RELEASE_IMAGE"), os.Getenv("NODE_ZERO_IP"))
	if err != nil {
		dialogs.PanicDialog(app, err)
	}
	validations.PrintConnectivityStatus(os.Stdout, false, true)

	if !validations.HasConnectivityIssue() {
		// Connectivity checks passed. Give 20 seconds for user
		// to start network configuration. If there is no input
		// application exits.
		exitAfterTimeout = true

		// view is the modal asking the user if they would still
		// like to change their network configuration.
		view := tview.NewModal().
			SetTextColor(tcell.ColorBlack).
			SetDoneFunc(userPromptHandler(app, pages, validations, exitAfterTimeout)).
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
		activateNetworkConfigurationScreen(app, pages, validations)
	}

	if err := app.SetRoot(pages, true).Run(); err != nil {
		dialogs.PanicDialog(app, err)
	}
}
