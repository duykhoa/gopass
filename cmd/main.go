package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/duykhoa/gopass/internal/gpg"

	// ...existing code...
	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/git"
)

// Removed local passwordStoreDir and gpgIdFile

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

func listPasswordEntries() ([]string, error) {
	dir := config.PasswordStoreDir()
	var entries []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".gpg") {
			rel, _ := filepath.Rel(dir, path)
			entries = append(entries, strings.TrimSuffix(rel, ".gpg"))
		}
		return nil
	})
	return entries, err
}

func main() {
	a := app.New()
	w := a.NewWindow("GoPass UI MVP")

	if !gpg.CheckGPGAvailable() {
		dialog.ShowError(fmt.Errorf("gpg command not found. Please install GnuPG (gpg) and restart the app."), w)
		w.ShowAndRun()
		return
	}

	status := widget.NewLabel("")

	var entries []string
	entries, _ = listPasswordEntries()
	selectedIdx := -1

	var decryptBtn *widget.Button
	entriesList := widget.NewList(
		func() int {
			entries, _ = listPasswordEntries()
			return len(entries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			entries, _ = listPasswordEntries()
			if i < len(entries) {
				o.(*widget.Label).SetText(entries[i])
			}
		},
	)
	entriesList.OnSelected = func(id widget.ListItemID) {
		selectedIdx = id
		status.SetText(fmt.Sprintf("Selected: %s", entries[id]))
		if decryptBtn != nil {
			decryptBtn.Enable()
		}
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		entriesList.Refresh()
		status.SetText("Entries refreshed")
	})

	syncBtn := widget.NewButton("Sync", func() {
		go func() {
			err := git.SyncWithRemote(config.PasswordStoreDir())
			if err != nil {
				dialog.ShowError(fmt.Errorf("git sync failed: %w", err), w)
			} else {
				dialog.ShowInformation("Git Sync", "Sync completed successfully", w)
			}
		}()
	})

	// Helper to show decrypted output in a selectable/copyable dialog
	showDecryptedDialog := func(parent fyne.Window, text string) {
		entry := widget.NewMultiLineEntry()
		entry.SetText(text)
		entry.Disable()
		entry.Wrapping = fyne.TextWrapWord
		content := container.NewVBox(entry)
		d := dialog.NewCustom("Decrypted", "OK", content, parent)
		d.Resize(fyne.NewSize(600, 400))
		d.Show()
	}

	decryptBtn = widget.NewButton("Decrypt", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), w)
			return
		}
		entry := entries[selectedIdx]
		gpgFile := filepath.Join(expandHome(config.PasswordStoreDir()), entry+".gpg")
		usr, _ := os.UserHomeDir()
		cachePath := filepath.Join(usr, ".gopass", "passphrase.cache")
		pass, valid, _ := gpg.DecryptCachedPassphrase(cachePath)
		if !valid {
			// Prompt for passphrase
			passDialog := widget.NewPasswordEntry()
			d := dialog.NewForm("Enter GPG Passphrase", "OK", "Cancel",
				[]*widget.FormItem{widget.NewFormItem("Passphrase", passDialog)},
				func(ok bool) {
					if !ok {
						return
					}
					pass = passDialog.Text
					// Cache for 30 minutes
					if err := gpg.EncryptAndCachePassphrase(pass, cachePath, 30*time.Minute); err != nil {
						fmt.Printf("[ERROR] Failed to cache passphrase: %v\n", err)
						dialog.ShowError(fmt.Errorf("Failed to cache passphrase: %w", err), w)
						// Proceed without caching
					}

					plaintext, err := gpg.DecryptGPGFileWithKey(gpgFile, pass)
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					showDecryptedDialog(w, plaintext)
				}, w)
			d.Show()
			return
		}
		plaintext, err := gpg.DecryptGPGFileWithKey(gpgFile, pass)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		showDecryptedDialog(w, plaintext)
	})
	decryptBtn.Disable()

	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		),
		fyne.NewMenu("Git",
			fyne.NewMenuItem("Sync", func() {
				go func() {
					err := git.SyncWithRemote(config.PasswordStoreDir())
					if err != nil {
						dialog.ShowError(fmt.Errorf("git sync failed: %w", err), w)
					} else {
						dialog.ShowInformation("Git Sync", "Sync completed successfully", w)
					}
				}()
			}),
		),
	)
	w.SetMainMenu(menu)

	entriesListScroll := container.NewVScroll(entriesList)
	entriesListScroll.SetMinSize(fyne.NewSize(0, 5*24))

	mainContent := container.NewVBox(
		container.NewHBox(refreshBtn, decryptBtn, syncBtn),
		widget.NewLabel("Password Entries:"),
		entriesListScroll,
	)

	content := container.NewBorder(nil, status, nil, nil, mainContent)
	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 600))
	w.ShowAndRun()
}
