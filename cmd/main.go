package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/git"
	"github.com/duykhoa/gopass/internal/gpg"
	"github.com/duykhoa/gopass/internal/service"
)

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

// Helper: show add/edit dialog with dynamic fields
func showAddOrEditDialog(w fyne.Window, title, okLabel, cancelLabel, entryName, templateName string, initialValues map[string]string, onSave func(entryName, templateName string, values map[string]string)) {
	templateNames := []string{service.TemplateFreeForm, service.TemplateEmailAndPassword}
	templateSelect := widget.NewSelect(templateNames, nil)
	if templateName != "" {
		templateSelect.SetSelected(templateName)
	} else {
		templateSelect.SetSelected(templateNames[0])
	}
	entryNameEntry := widget.NewEntry()
	if entryName != "" {
		entryNameEntry.SetText(entryName)
		entryNameEntry.Disable()
	}
	entryNameEntry.SetPlaceHolder("Entry name (e.g. github)")
	fieldWidgets := map[string]*widget.Entry{}
	c := cases.Title(language.English)
	form := widget.NewForm(
		widget.NewFormItem("Entry Name", entryNameEntry),
		widget.NewFormItem("Template", templateSelect),
	)
	var updateFields func(string)
	updateFields = func(templateName string) {
		// Keep the first two items (Entry Name, Template), replace the rest
		items := form.Items[:2]
		fieldWidgets = map[string]*widget.Entry{}
		tmpl := service.GetTemplateByName(templateName)
		if tmpl != nil {
			for _, field := range tmpl.Fields {
				entry := widget.NewEntry()
				if field == "content" || field == "extra" {
					entry.MultiLine = true
				}
				if initialValues != nil {
					entry.SetText(initialValues[field])
				}
				fieldWidgets[field] = entry
				items = append(items, widget.NewFormItem(c.String(field), entry))
			}
		}
		form.Items = items
		form.Refresh()
	}
	updateFields(templateSelect.Selected)
	templateSelect.OnChanged = func(name string) {
		updateFields(name)
	}
	d := dialog.NewCustom(title, "Submit", form, w)
	form.OnSubmit = func() {
		entryName := entryNameEntry.Text
		templateName := templateSelect.Selected
		values := map[string]string{}
		for k, entry := range fieldWidgets {
			values[k] = entry.Text
		}
		d.Hide()
		onSave(entryName, templateName, values)
	}
	d.SetDismissText(cancelLabel)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// Helper: show edit dialog with fields from decrypted content
func showEditDialogWithContent(w fyne.Window, entryName, content string, onSave func(values map[string]string)) {
	templateName := getTemplateFromContent(content)
	tmpl := service.GetTemplateByName(templateName)
	values := parseFieldsFromContent(content, tmpl)
	fieldWidgets := map[string]*widget.Entry{}
	var formItems []*widget.FormItem
	c := cases.Title(language.English)
	for _, field := range tmpl.Fields {
		entry := widget.NewEntry()
		if field == "content" || field == "extra" {
			entry.MultiLine = true
		}
		entry.SetText(values[field])
		fieldWidgets[field] = entry
		formItems = append(formItems, widget.NewFormItem(c.String(field), entry))
	}
	d := dialog.NewForm("Edit Entry", "Save", "Cancel", formItems, func(ok bool) {
		if !ok {
			return
		}
		newValues := map[string]string{}
		for k, entry := range fieldWidgets {
			newValues[k] = entry.Text
		}
		onSave(newValues)
	}, w)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// Helper: parse template name from content
func getTemplateFromContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "template:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "template:"))
		}
	}
	return service.TemplateFreeForm
}

// Helper: parse fields from content
func parseFieldsFromContent(content string, tmpl *service.Template) map[string]string {
	values := map[string]string{}
	lines := strings.Split(content, "\n")
	for _, field := range tmpl.Fields {
		for _, line := range lines {
			if strings.HasPrefix(line, field+":") {
				values[field] = strings.TrimSpace(strings.TrimPrefix(line, field+":"))
			}
		}
	}
	return values
}

func main() {
	a := app.New()
	w := a.NewWindow("GoPass UI MVP")

	if !gpg.CheckGPGAvailable() {
		dialog.ShowError(fmt.Errorf("gpg command not found. Please install GnuPG (gpg) and restart the app"), w)
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
		pass, valid := service.GetCachedPassphrase()
		if !valid {
			passDialog := widget.NewPasswordEntry()
			d := dialog.NewForm("Enter GPG Passphrase", "OK", "Cancel",
				[]*widget.FormItem{widget.NewFormItem("Passphrase", passDialog)},
				func(ok bool) {
					if !ok {
						return
					}
					pass = passDialog.Text
					// Try decryption first, only cache if successful
					result := service.DecryptAndMaybeCacheWithCache(entry, pass, false)
					if result.Err != nil {
						dialog.ShowError(result.Err, w)
						return
					}
					// Now cache passphrase
					_ = service.DecryptAndMaybeCacheWithCache(entry, pass, true)
					showDecryptedDialog(w, result.Plaintext)
				}, w)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
		result := service.DecryptAndMaybeCacheWithCache(entry, pass, false)
		if result.Err != nil {
			dialog.ShowError(result.Err, w)
			return
		}
		showDecryptedDialog(w, result.Plaintext)
	})
	decryptBtn.Disable()

	addEntryBtn := widget.NewButton("Add Entry", func() {
		showAddOrEditDialog(w, "Add New Entry", "Submit", "Cancel", "", "", nil, func(entryName, templateName string, values map[string]string) {
			// gpgId := config.GPGId()
			// home, err := os.UserHomeDir()
			// if err != nil {
			// 	dialog.ShowError(fmt.Errorf("Failed to get user home directory: %w", err), w)
			// 	return
			// }
			// armoredKeyPath := filepath.Join(home, ".gopass", gpgId+".public.asc")
			// if _, err := os.Stat(armoredKeyPath); err != nil {
			// 	dialog.ShowError(fmt.Errorf("No armored public key file found for recipient '%s' at %s. Please ensure the correct GPG ID is set in your config and the armored key exists (should be named <gpg_id>.public.asc).", gpgId, armoredKeyPath), w)
			// 	return
			// }
			req := service.AddEditRequest{
				EntryName:    entryName,
				TemplateName: templateName,
				Fields:       values,
				// GPGId:        gpgId,
			}
			result := service.AddOrEditEntry(req)
			if result.Err != nil {
				dialog.ShowError(result.Err, w)
				return
			}
			entriesList.Refresh()
			status.SetText(service.GetGitStatus())
		})
	})

	// Delete Entry Button
	deleteEntryBtn := widget.NewButton("Delete Entry", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), w)
			return
		}
		entry := entries[selectedIdx]
		dialog.ShowConfirm("Delete Entry", fmt.Sprintf("Are you sure you want to delete '%s'?", entry), func(ok bool) {
			if !ok {
				return
			}
			// Remove file
			path := filepath.Join(config.PasswordStoreDir(), entry+".gpg")
			if err := os.Remove(path); err != nil {
				dialog.ShowError(fmt.Errorf("failed to delete entry: %w", err), w)
				return
			}
			entriesList.Refresh()
			status.SetText(fmt.Sprintf("Entry '%s' deleted", entry))
		}, w)
	})

	// Edit Entry Button
	editEntryBtn := widget.NewButton("Edit Entry", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), w)
			return
		}
		entry := entries[selectedIdx]
		// Decrypt entry to get fields
		pass, valid := service.GetCachedPassphrase()
		if !valid {
			passDialog := widget.NewPasswordEntry()
			d := dialog.NewForm("Enter GPG Passphrase", "OK", "Cancel",
				[]*widget.FormItem{widget.NewFormItem("Passphrase", passDialog)},
				func(ok bool) {
					if !ok {
						return
					}
					pass = passDialog.Text
					result := service.DecryptAndMaybeCacheWithCache(entry, pass, false)
					if result.Err != nil {
						dialog.ShowError(result.Err, w)
						return
					}
					_ = service.DecryptAndMaybeCacheWithCache(entry, pass, true)
					showEditDialogWithContent(w, entry, result.Plaintext, func(values map[string]string) {
						req := service.AddEditRequest{
							EntryName:    entry,
							TemplateName: getTemplateFromContent(result.Plaintext),
							Fields:       values,
							GPGId:        config.GPGId(),
						}
						res := service.AddOrEditEntry(req)
						if res.Err != nil {
							dialog.ShowError(res.Err, w)
							return
						}
						entriesList.Refresh()
						status.SetText("Entry updated. Please sync to remote.")
						dialog.ShowInformation("Sync Reminder", "Entry updated. Please sync to remote.", w)
					})
				}, w)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
		result := service.DecryptAndMaybeCacheWithCache(entry, pass, false)
		if result.Err != nil {
			dialog.ShowError(result.Err, w)
			return
		}
		showEditDialogWithContent(w, entry, result.Plaintext, func(values map[string]string) {
			req := service.AddEditRequest{
				EntryName:    entry,
				TemplateName: getTemplateFromContent(result.Plaintext),
				Fields:       values,
				GPGId:        config.GPGId(),
			}
			res := service.AddOrEditEntry(req)
			if res.Err != nil {
				dialog.ShowError(res.Err, w)
				return
			}
			entriesList.Refresh()
			status.SetText("Entry updated. Please sync to remote.")
			dialog.ShowInformation("Sync Reminder", "Entry updated. Please sync to remote.", w)
		})
	})

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
	entriesListScroll.SetMinSize(fyne.NewSize(0, 10*24))

	mainContent := container.NewVBox(
		container.NewHBox(refreshBtn, decryptBtn, syncBtn, addEntryBtn, editEntryBtn, deleteEntryBtn),
		widget.NewLabel("Password Entries:"),
		entriesListScroll,
	)

	content := container.NewBorder(nil, status, nil, nil, mainContent)
	w.SetContent(content)
	w.Resize(fyne.NewSize(600, 700))
	w.ShowAndRun()
}
