package pico

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/service"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MsgType string

const (
	MsgType_UpdateStatus = "MsgTypeUpdateStatus"
)

type Msg struct {
	Type    MsgType
	Content string
}

type controller struct {
	View    *view
	Model   *model
	quit    chan struct{}
	msgChan chan Msg
}

func (c *controller) CheckAndRender() {
	_, err := os.Stat(config.PasswordStoreDir())

	if err != nil {
		slog.Error("TODO: password store dir doesn't exist")

		return
	}

	c.ShowMainPage()
}

func (c *controller) ShowMainPage() {
	c.View.ShowPage("main")

	entries, err := service.ListPasswordEntries(config.PasswordStoreDir())

	if err != nil {
		slog.Error("Failed to load password entries", slog.Any("error", err))
		return
	}

	slog.Info("Entries list", slog.Any("data", entries))
}

func (c *controller) ListenToEvents() error {
	for {
		select {
		case <-c.quit:
			slog.Debug("Controller receives quit signal, return nil")
			time.Sleep(300 * time.Millisecond) // Give some time for UI to show the last status update
			c.View.app.Stop()
			return nil
		case msg := <-c.msgChan:
			slog.Info("Receive msg from msgChan", slog.Any("msg", msg))

			switch msg.Type {
			case MsgType_UpdateStatus:
				c.View.SetStatusText(msg.Content)

				close(c.quit)
			}
		}
	}

	return nil
}

type view struct {
	app             *tview.Application
	pages           *tview.Pages
	msgChan         chan Msg
	passwordEntries *tview.TextView
	passwordDetail  *tview.TextView
	statusText      *tview.TextView
}

func (v *view) Render() error {
	return v.app.Run()
}

func (v *view) Init() {
	headerLine := tview.NewTextView().SetTextAlign(tview.AlignRight)
	addMenu := fmt.Sprintf("%sdd", WrapColor("A", "#ff0000"))
	syncMenu := fmt.Sprintf("%sync", WrapColor("S", "#ff0000"))
	quitMenu := fmt.Sprintf("%suit", WrapColor("Q", "#ff0000"))
	helpMenu := fmt.Sprintf("%selp", WrapColor("H", "#ff0000"))

	statusText := tview.NewTextView().SetText(WrapColor("Password entries are loaded, have a good day!", "blue")).
		SetTextAlign(tview.AlignLeft).SetDynamicColors(true)
	v.statusText = statusText

	headerLine.SetText(
		fmt.Sprintf("%s\t%s\t%s\t%s\t", addMenu, syncMenu, quitMenu, helpMenu),
	).SetDynamicColors(true)

	passEntries := tview.NewTextView().SetText("Password Entries")
	v.passwordEntries = passEntries

	passDetail := tview.NewTextView().SetText("Password Detail")
	v.passwordDetail = passDetail

	grid := tview.NewGrid().SetColumns(40, 0).SetRows(2, 0, 1).SetBorders(true)
	grid.AddItem(headerLine, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(passEntries, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(passDetail.SetTextAlign(tview.AlignLeft), 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(statusText, 2, 0, 1, 2, 0, 0, false)

	v.pages = tview.NewPages()

	v.pages.AddPage("main", grid, false, true)

	v.app.SetRoot(grid, true)

	v.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlA:
			v.statusText.SetText("user press a")
		case tcell.KeyCtrlS:
			v.statusText.SetText("user press s")
		case tcell.KeyCtrlQ:
			v.msgChan <- Msg{
				Type:    MsgType_UpdateStatus,
				Content: "User presses CtrlQ, exiting!",
			}
		case tcell.KeyCtrlH:
			v.statusText.SetText("user press h")
		}

		return event
	})
}

func (v *view) ShowPage(name string) {
	v.pages.ShowPage(name)
}

func (v *view) SetStatusText(status string) {
	v.statusText.SetText(status)
}

type model struct {
	Entries struct{}
}

type app struct {
	Controller *controller
}

func (a *app) Run() error {
	var err error

	a.Controller.View.Init()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err = a.Controller.View.Render()

		if err != nil {
			slog.Error("Error in Render UI", slog.Any("error", err))
		}
	}()

	go func() {
		defer wg.Done()

		a.Controller.CheckAndRender()
		err = a.Controller.ListenToEvents()
		if err != nil {
			slog.Error("Error in ListenToEvents", slog.Any("error", err))
		}
	}()

	wg.Wait()
	slog.Debug("Exiting the app")

	return err
}

func NewPico() *app {
	msgChan := make(chan Msg)
	quit := make(chan struct{})

	view := &view{
		msgChan: msgChan,
		app:     tview.NewApplication(),
	}

	model := &model{}

	controller := &controller{
		View:    view,
		Model:   model,
		quit:    quit,
		msgChan: msgChan,
	}

	a := &app{
		Controller: controller,
	}

	return a
}

func WrapColor(str string, color string) string {
	return fmt.Sprintf("[%s]%s[white]", color, str)
}
