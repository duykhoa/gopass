package pico

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

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
				statusLine, err := c.View.GetComponentByID("status")
				statusText := statusLine.(*tview.TextView)

				if err != nil {
					return err
				}

				statusText.SetText(msg.Content)

				close(c.quit)
			}
		}
	}
}

type view struct {
	app        *tview.Application
	components map[string]tview.Primitive
	msgChan    chan Msg
}

func WrapColor(str string, color string) string {
	return fmt.Sprintf("[%s]%s[white]", color, str)
}

func (v *view) RegisterComponent(id string, comp tview.Primitive) error {
	if _, ok := v.components[id]; ok {
		return fmt.Errorf("component with id: %s is existed", id)
	}

	v.components[id] = comp

	return nil
}

func (v *view) GetComponentByID(id string) (tview.Primitive, error) {
	comp, ok := v.components[id]

	if !ok {
		return nil, fmt.Errorf("component with ID: %s doesn't exist", id)
	}

	return comp, nil
}

func (v *view) Render() error {
	var err error

	headerLine := tview.NewTextView().SetTextAlign(tview.AlignRight)
	addMenu := fmt.Sprintf("%sdd", WrapColor("A", "#ff0000"))
	syncMenu := fmt.Sprintf("%sync", WrapColor("S", "#ff0000"))
	quitMenu := fmt.Sprintf("%suit", WrapColor("Q", "#ff0000"))
	helpMenu := fmt.Sprintf("%selp", WrapColor("H", "#ff0000"))

	statusText := tview.NewTextView().SetText(WrapColor("Password entries are loaded, have a good day!", "blue")).
		SetTextAlign(tview.AlignLeft).SetDynamicColors(true)
	v.RegisterComponent("status", statusText)

	headerLine.SetText(
		fmt.Sprintf("%s\t%s\t%s\t%s\t", addMenu, syncMenu, quitMenu, helpMenu),
	).SetDynamicColors(true)

	passEntries := tview.NewTextView().SetText("PasswordEntries")
	v.RegisterComponent("password_entries", passEntries)

	passDetail := tview.NewTextView().SetText("PasswordDetail")
	v.RegisterComponent("password_detail", passDetail)

	grid := tview.NewGrid().SetColumns(40, 0).SetRows(2, 0, 1).SetBorders(true)
	grid.AddItem(headerLine, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(passEntries, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(passDetail.SetTextAlign(tview.AlignLeft), 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(statusText, 2, 0, 1, 2, 0, 0, false)

	pages := tview.NewPages()

	pages.AddPage("main", grid, false, true)
	pages.ShowPage("main")

	v.app.SetRoot(grid, true)

	v.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		statusLine, err := v.GetComponentByID("status")
		statusText := statusLine.(*tview.TextView)

		if err != nil {
			slog.Error("failed to get component status")
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlA:
			statusText.SetText("user press a")
		case tcell.KeyCtrlS:
			statusText.SetText("user press s")
		case tcell.KeyCtrlQ:
			v.msgChan <- Msg{
				Type:    MsgType_UpdateStatus,
				Content: "User presses CtrlQ, exiting!",
			}
		case tcell.KeyCtrlH:
			statusText.SetText("user press h")
		}

		return event
	})

	err = v.app.Run()

	return err
}

type model struct {
}

type app struct {
	Controller *controller
}

func (a *app) Run() error {
	var err error

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
		msgChan:    msgChan,
		app:        tview.NewApplication(),
		components: make(map[string]tview.Primitive),
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
