package agent_tui

import (
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/dialogs"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/ui"
	"github.com/rivo/tview"
)

func App(app *tview.Application, config checks.Config) {
	var appUI *ui.UI
	if app == nil {
		appUI = ui.NewUI(config)
		app = appUI.GetApp()
	}
	controller := ui.NewController(appUI)

	engine := checks.NewEngine(controller.GetChan(), config)

	controller.Init()
	engine.Init()

	if err := app.Run(); err != nil {
		dialogs.PanicDialog(app, err)
	}
}
