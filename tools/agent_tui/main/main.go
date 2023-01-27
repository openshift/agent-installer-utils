package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/forms"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	QUIT      string = "Quit"
	CONFIGURE string = "Configure"
	YES       string = "Yes"
	NO        string = "No"
)

func node0Handler(app *tview.Application, pages *tview.Pages) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		if buttonLabel == YES {
			// TODO: Print addressing, offer to configure, done

			node0Form := forms.Node0Form(app, pages)
			pages.AddPage("node0Form", node0Form, true, true)
		} else {
			regNodeForm := forms.RegNodeModalForm(app, pages)
			pages.AddPage("regNodeConfig", regNodeForm, true, true)
		}
	}
}

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	background := tview.NewBox().
		SetBorder(false).
		SetBackgroundColor(newt.ColorBlue)

	node0 := tview.NewModal().
		SetText("Do you wish for this node to be the one that runs the installation service (only one node may perform this function)?").
		SetTextColor(tcell.ColorBlack).
		SetDoneFunc(node0Handler(app, pages)).
		SetBackgroundColor(newt.ColorGray).
		SetButtonTextColor(tcell.ColorBlack).
		SetButtonBackgroundColor(tcell.ColorDarkGray)

	node0Buttons := []string{YES, NO}
	node0.AddButtons(node0Buttons)

	pages.AddPage("background", background, true, true).
		AddPage("Node0", node0, true, true)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
