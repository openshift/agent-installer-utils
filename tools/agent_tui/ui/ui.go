package ui

import (
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type UI struct {
	app                 *tview.Application
	pages               *tview.Pages
	mainFlex, innerFlex *tview.Flex
	primaryCheck        *tview.Table
	checks              *tview.Table    // summary of all checks
	details             *tview.TextView // where errors from checks are displayed
	netConfigForm       *tview.Form     // contains "Configure network" button
	timeoutModal        *tview.Modal    // popup window that times out
	splashScreen        *tview.Modal    // display initial waiting message
	nmtuiActive         atomic.Value
	timeoutDialogActive atomic.Value
	timeoutDialogCancel chan bool
	dirty               atomic.Value // dirty flag set if the user interacts with the ui

	// Rendezvous node IP workflow
	rendezvousIPForm            *tview.Form
	selectIPForm                *tview.Form
	selectIPList                *tview.List
	errorModal                  *tview.Modal
	rendezvousIPFormActive      atomic.Value
	rendezvousIPSaveSuccessModal *tview.Modal

	focusableItems []tview.Primitive // the list of widgets that can be focused
	focusedItem    int               // the current focused widget

	logger *logrus.Logger
}

func NewUI(app *tview.Application, config checks.Config, logger *logrus.Logger) *UI {
	ui := &UI{
		app:                 app,
		timeoutDialogCancel: make(chan bool),
		logger:              logger,
	}
	ui.nmtuiActive.Store(false)
	ui.timeoutDialogActive.Store(false)
	ui.rendezvousIPFormActive.Store(false)
	ui.dirty.Store(false)
	ui.create(config)
	return ui
}

func (u *UI) GetApp() *tview.Application {
	return u.app
}

func (u *UI) GetPages() *tview.Pages {
	return u.pages
}

func (u *UI) setFocusToChecks() {
	// reset u.focusableItems to those on the checks page
	u.focusableItems = []tview.Primitive{
		u.netConfigForm.GetButton(0),
		u.netConfigForm.GetButton(1),
	}
	u.pages.SwitchToPage(PAGE_CHECKSCREEN)
	// shifting focus back to the "Configure network"
	// button requires setting focus in this sequence
	// form -> form-button
	u.app.SetFocus(u.netConfigForm)
	u.app.SetFocus(u.netConfigForm.GetButton(0))
}

func (u *UI) setFocusToRendezvousIP() {
	u.setIsRendezousIPFormActive(true)
	// reset u.focusableItems to those on the rendezvous IP page
	u.focusableItems = []tview.Primitive{
		u.rendezvousIPForm.GetFormItemByLabel(FIELD_ENTER_RENDEZVOUS_IP),
		u.selectIPForm.GetButton(0),
		u.rendezvousIPForm.GetButton(0),
	}

	u.pages.SwitchToPage(PAGE_RENDEZVOUS_IP)
	u.app.SetFocus(u.rendezvousIPForm)
	u.app.SetFocus(u.rendezvousIPForm.GetFormItemByLabel(FIELD_ENTER_RENDEZVOUS_IP))
}

func (u *UI) setFocusToSelectIP() {
	u.setIsRendezousIPFormActive(true)
	u.pages.SwitchToPage(PAGE_SET_NODE_AS_RENDEZVOUS)

	u.app.SetFocus(u.selectIPList)
}

func (u *UI) IsNMTuiActive() bool {
	return u.nmtuiActive.Load().(bool)
}

func (u *UI) setIsTimeoutDialogActive(isActive bool) {
	u.timeoutDialogActive.Store(isActive)
}

func (u *UI) IsTimeoutDialogActive() bool {
	return u.timeoutDialogActive.Load().(bool)
}

func (u *UI) setIsRendezousIPFormActive(isActive bool) {
	u.rendezvousIPFormActive.Store(isActive)
}

func (u *UI) IsRendezvousIPFormActive() bool {
	return u.rendezvousIPFormActive.Load().(bool)
}

func (u *UI) IsDirty() bool {
	return u.dirty.Load().(bool)
}

func (u *UI) create(config checks.Config) {
	u.pages = tview.NewPages()
	u.createCheckPage(config)
	u.createTimeoutModal(config)
	u.createSplashScreen()
	u.createRendezvousIPPage(config)
	u.createErrorModal()
	u.createSelectHostIPPage()
	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !u.IsRendezvousIPFormActive() {
			// Any interaction with the rendezvous IP form does
			// not count as dirtying the UI. This prevents
			// interactions with the rendezvous IP form from
			// preventing the timeout dialog from appearing
			// if the release image check passes.
			u.dirty.Store(true)
		}
		return event
	})
}
