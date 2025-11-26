package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
)

type App struct {
	Window  fyne.Window
	Screens *Screens
}

func (a *App) ShowScreen(s string) {
	screenFunc, ok := a.Screens.GetScreen(s)
	if !ok {
		ShowErrorDialog(a.Window, fmt.Errorf("screen not found: %s", s))
		return
	}

	content := screenFunc(a)
	a.Window.SetContent(content)
	a.Window.SetTitle(fmt.Sprintf("GoPass - %s", s))
	a.Window.Resize(fyne.NewSize(600, 500))
}

type ScreenFunction = func(*App) fyne.CanvasObject

type Screens struct {
	Screens map[string]ScreenFunction
}

func NewScreens() *Screens {
	return &Screens{Screens: map[string]ScreenFunction{}}
}

func (s *Screens) GetScreen(name string) (ScreenFunction, bool) {
	screenFunc, ok := s.Screens[name]
	return screenFunc, ok
}

func (s *Screens) AddScreen(name string, screenFunc ScreenFunction) {
	// just overwrite if exists
	s.Screens[name] = screenFunc
}
