package agent_tui

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

const (
	waitTimeout = 5 * time.Second
)

// AppTester is a test helper class to allow writing integration tests.
// It works by injecting a SimulationScreen, a normal Screen object with
// some additional testing features.
// The class offers some methods to interact with the Agent TUI screen, so
// that it would be possible to mimick exactly the user actions required to
// accomplish a given activity.
// In addition, a number of verification methods are present to check the results
// of the action performed (WaitFor...)
type AppTester struct {
	t      *testing.T
	screen tcell.SimulationScreen
	app    *tview.Application
}

// Creates a new instance of AppTester
func NewAppTester(t *testing.T, debug ...bool) *AppTester {

	s := tcell.NewSimulationScreen("")
	if s == nil {
		t.Fatalf("Unable to create simulation screen")
	}
	if err := s.Init(); err != nil {
		t.Fatalf("Failed to initialize screen: %v", err)
	}

	return &AppTester{
		t:      t,
		screen: s,
		app:    tview.NewApplication().SetScreen(s),
	}
}

// Starts a new Agent TUI in background
func (a *AppTester) Start() {
	go App(a.app)
}

// Releases all the resources and stop the app
func (a *AppTester) Stop() {
	a.app.Stop()
}

// Waits until a button with the specified label becomes focused
func (a *AppTester) WaitForButton(expectedLabel string) {
	a.waitForItemWithLabel(expectedLabel, func(p tview.Primitive) string {
		item, ok := p.(*tview.Button)
		if !ok {
			return ""
		}
		return item.GetLabel()
	})
}

// Waits until a button with the specified label becomes focused
func (a *AppTester) WaitForInputField(expectedLabel string) {
	a.waitForItemWithLabel(expectedLabel, func(p tview.Primitive) string {
		item, ok := p.(*tview.InputField)
		if !ok {
			return ""
		}
		return item.GetLabel()
	})
}

func (a *AppTester) waitForItemWithLabel(expectedLabel string, labelGetter func(p tview.Primitive) string) {
	a.t.Helper()
	ok := assert.Eventually(a.t, func() bool {
		p := a.app.GetFocus()
		if p == nil {
			return false
		}
		actualLabel := labelGetter(p)
		return expectedLabel == actualLabel
	}, waitTimeout, 10*time.Millisecond)

	if !ok {
		assert.FailNow(a.t, fmt.Sprintf("widget with label '%s' not found", expectedLabel))
	}
}

// Moves the current cursor to the right
func (a *AppTester) ScreenMoveCursorRight() {
	a.screenPressKey(tcell.KeyRight)
}

// Press the Tab key
func (a *AppTester) ScreenPressTab() {
	a.screenPressKey(tcell.KeyTAB)
}

// Press the Enter key
func (a *AppTester) ScreenPressEnter() {
	a.screenPressKey(tcell.KeyEnter)
}

func (a *AppTester) screenPressKey(key tcell.Key) {
	a.screen.InjectKey(key, rune(0), tcell.ModNone)
	time.Sleep(1 * time.Millisecond)
}

// Types a string at the current screen position and then press enter
func (a *AppTester) ScreenTypeText(text string) {
	for _, c := range text {
		a.screen.InjectKey(tcell.KeyRune, rune(c), tcell.ModNone)
		time.Sleep(1 * time.Millisecond)
	}
}

func (a *AppTester) fetchScreenContent() []string {
	cells, w, h := a.screen.GetContents()
	lines := []string{}

	for y := 0; y < h; y++ {
		line := ""
		for x := 0; x < w; x++ {
			r := cells[x+y*w].Runes[0]
			if !unicode.IsSymbol(r) {
				line += string(r)
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

// Wait until the current screen buffer contains the specified labels, or timeout
func (a *AppTester) WaitForScreenContent(labels ...string) {
	a.t.Helper()
	ok := assert.Eventually(a.t, func() bool {
		lines := a.fetchScreenContent()
		for _, label := range labels {
			found := false
			for _, l := range lines {
				if strings.Contains(l, label) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}, waitTimeout, 10*time.Millisecond)

	if !ok {
		assert.FailNow(a.t, fmt.Sprintf("Screen does not contain '%s'", labels))
	}
}

// Print the content of the current screen to the terminal in a raw format.
// Just useful for debugging.
func (a *AppTester) DumpScreen() {
	cells, w, h := a.screen.GetContents()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[x+y*w]
			fmt.Print(string(c.Bytes))
		}
		fmt.Println()
	}
}
