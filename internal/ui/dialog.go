package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// Helper: show error dialog
func ShowErrorDialog(w fyne.Window, err error) {
	d := dialog.NewError(err, w)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}