package agent_tui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/dialogs"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/forms"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

func App(app *tview.Application) {
	if app == nil {
		app = tview.NewApplication()
	}
	pages := tview.NewPages()

	background := tview.NewBox().
		SetBorder(false).
		SetBackgroundColor(newt.ColorBlue)

	pages.AddPage("background", background, true, true).
		AddPage("Node0", forms.IsNode0Modal(app, pages), true, true)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		dialogs.PanicDialog(app, err)
	}
}
