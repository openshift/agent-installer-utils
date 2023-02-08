package forms

import (
	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/net"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/newt"
	"github.com/rivo/tview"
)

const (
	CONNECTIVITYCHECK   string = "Check Connectivity"
	NETCONFIGURE        string = "Configure Networking"
	DONE                string = "Done"
	RENDEZVOUSLABEL     string = "Rendezvous IP Address"
	RELEASE_IMAGE_LABEL string = "Release Image"
)

func RegNodeModalForm(app *tview.Application, pages *tview.Pages, validations *net.Validations) tview.Primitive {
	statusView := tview.NewTextView()
	statusView.SetBackgroundColor(newt.ColorGray).
		SetBorder(true).
		SetBorderColor(newt.ColorBlue).
		SetTitle("Status").
		SetTitleColor(tcell.ColorBlack)
	statusView.SetTextColor(tcell.ColorBlack).
		SetDynamicColors(true)
	statusView.SetScrollable(true).SetWrap(true)

	releaseImageTextView := tview.NewTextView()
	releaseImageTextView.SetLabel("[black]" + RELEASE_IMAGE_LABEL + ": " + validations.ReleaseImageURL)
	releaseImageTextView.SetTextColor(tcell.ColorBlack)
	releaseImageTextView.SetDynamicColors(true)
	releaseImageTextView.SetBackgroundColor(newt.ColorGray)

	//AddTextView(RELEASE_IMAGE_LABEL+": "+validations.ReleaseImageURL, "", 70, 1, true, false).
	// AddTextView(RENDEZVOUSLABEL+": "+validations.RendezvousHostIP, "", 70, 1, true, false).
	rendezvousIPTextView := tview.NewTextView()
	rendezvousIPTextView.SetLabel("[black]" + RENDEZVOUSLABEL + ": " + validations.RendezvousHostIP).
		SetTextColor(tcell.ColorBlack).
		SetBackgroundColor(newt.ColorGray)

	checkButton := tview.NewButton(CONNECTIVITYCHECK).
		SetSelectedFunc(func() {
			statusView.Clear()
			go func() {
				validations.PrintConnectivityStatus(statusView, false, false)
				app.Draw()
			}()
		}).
		SetBackgroundColor(newt.ColorGray)

	configureButton := tview.NewButton(NETCONFIGURE).
		SetSelectedFunc(net.NMTUIRunner(app, pages, nil)).
		SetBackgroundColor(newt.ColorGray)

	doneButton := tview.NewButton(DONE).
		SetSelectedFunc(func() {
			if !validations.HasConnectivityIssue() {
				app.Stop()
			} else {
				statusView.SetText("[red::b]Can't continue installation without a successful connectivity check")
			}
		}).
		SetBackgroundColor(newt.ColorGray)

	regNodeConfigForm := tview.NewForm()
	// AddFormItem(releaseImageTextView).
	// AddFormItem(rendezvousIPTextView).
	// AddButton(CONNECTIVITYCHECK, func() {
	// 	statusView.Clear()
	// 	go func() {
	// 		validations.PrintConnectivityStatus(statusView, false, false)
	// 		app.Draw()
	// 	}()
	// }).
	// AddButton(NETCONFIGURE, net.NMTUIRunner(app, pages, nil)).
	// AddButton(DONE, func() {
	// 	if !validations.HasConnectivityIssue() {
	// 		app.Stop()
	// 	} else {
	// 		statusView.SetText("[red::b]Can't continue installation without a successful connectivity check")
	// 	}
	// })
	regNodeConfigForm.
		SetLabelColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorBlack).
		SetTitle("Agent-based Installer Network Connectivity Check").
		SetTitleColor(tcell.ColorBlack).
		SetBackgroundColor(newt.ColorGray)

	// Prefill the status view if the initial validation checks performed when
	// the application started up indicated there is an issue.
	if validations.HasConnectivityIssue() {
		validations.PrintConnectivityStatus(statusView, true, false)
	}

	// Change navigation. By default Tab moves through form buttons. Now,
	// * Left and Right keys moves through buttons in the form
	// * Tab and Back Tab moves to status view
	regNodeConfigForm.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			if _, index := regNodeConfigForm.GetFocusedItemIndex(); index == (regNodeConfigForm.GetButtonCount() - 1) {
				app.SetFocus(statusView)
				eventKey = nil
			} else {
				app.SetFocus(regNodeConfigForm.GetButton(index + 1))
				eventKey = event
			}
		case tcell.KeyBacktab, tcell.KeyLeft:
			if _, index := regNodeConfigForm.GetFocusedItemIndex(); index == 0 {
				app.SetFocus(statusView)
				eventKey = nil
			} else {
				app.SetFocus(regNodeConfigForm.GetButton((index) - 1))
				eventKey = event
			}
		default:
			eventKey = event
		}
		return
	})

	// Register tab keys to switch to connectivity check form
	statusView.SetInputCapture(func(event *tcell.EventKey) (eventKey *tcell.EventKey) {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyRight:
			app.SetFocus(regNodeConfigForm)
			eventKey = nil
		default:
			eventKey = event
		}
		return
	})

	// grid := tview.NewGrid().
	// 	AddItem(releaseImageTextView, 0, 0, 0, 0, 0, 0, false).
	// 	AddItem(rendezvousIPTextView, 2, 1, 0, 0, 0, 0, false).
	// 	AddItem(checkButton, 3, 1, 0, 0, 0, 0, false).
	// 	AddItem(configureButton, 3, 2, 0, 0, 0, 0, false).
	// 	AddItem(doneButton, 3, 3, 0, 0, 0, 0, false).
	// 	AddItem(statusView, 4, 1, 0, 0, 0, 0, false).
	// 	SetRows(1, 1, 1, 1, 10, 1).
	// 	SetColumns(20, 20, 20, 20).
	// 	SetBorder(true).
	// 	SetBorderColor(tcell.ColorBlue).
	// 	SetTitle("Agent installer network boot setup").
	// 	SetTitleColor(tcell.ColorBlack).
	// 	SetBackgroundColor(newt.ColorGray)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(releaseImageTextView, 0, 1, false).
		AddItem(rendezvousIPTextView, 0, 1, false).
		AddItem(checkButton, 0, 1, false).
		AddItem(configureButton, 0, 1, false).
		AddItem(doneButton, 0, 1, false).
		AddItem(statusView, 0, 10, false)

	flex.SetBorder(true).
		SetBorderColor(newt.ColorBlue).
		SetTitle("Agent installer network boot setup").
		SetTitleColor(tcell.ColorBlack).
		SetBackgroundColor(newt.ColorGray)

	width := 80
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(flex, width, 1, true).
		AddItem(nil, 0, 1, false)
}
