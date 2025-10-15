package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/ui"

	"github.com/duykhoa/gopass/internal/git"
	"github.com/duykhoa/gopass/internal/gpg"
	"github.com/duykhoa/gopass/internal/service"
	"github.com/duykhoa/gopass/internal/store"
)

func listPasswordEntries(dir string) ([]string, error) {
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
	var showDialog func(selectedTemplate string, entryNameValue string)
	showDialog = func(selectedTemplate string, entryNameValue string) {
		templateNames := []string{service.TemplateFreeForm, service.TemplateEmailAndPassword}
		templateSelect := widget.NewSelect(templateNames, nil)
		templateSelect.SetSelected(selectedTemplate)
		entryNameEntry := widget.NewEntry()
		if entryName != "" && entryNameValue == "" {
			entryNameEntry.SetText(entryName)
			entryNameEntry.Disable()
		} else if entryNameValue != "" {
			entryNameEntry.SetText(entryNameValue)
		}
		entryNameEntry.SetPlaceHolder("Entry name (e.g. github)")
		fieldWidgets := map[string]*widget.Entry{}
		c := cases.Title(language.English)
		var formItems []*widget.FormItem
		formItems = append(formItems, widget.NewFormItem("Entry Name", entryNameEntry))
		formItems = append(formItems, widget.NewFormItem("Template", templateSelect))
		tmpl := service.GetTemplateByName(selectedTemplate)
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
				formItems = append(formItems, widget.NewFormItem(c.String(field), entry))
			}
		}
		d := dialog.NewForm(title, okLabel, cancelLabel, formItems, func(ok bool) {
			if !ok {
				return
			}
			entryNameVal := entryNameEntry.Text
			templateName := templateSelect.Selected
			values := map[string]string{}
			for k, entry := range fieldWidgets {
				values[k] = entry.Text
			}
			onSave(entryNameVal, templateName, values)
		}, w)
		d.Resize(fyne.NewSize(500, 400))
		d.Show()
		templateSelect.OnChanged = func(name string) {
			d.Hide()
			showDialog(name, entryNameEntry.Text)
		}
	}
	// Start with initial template
	if templateName != "" {
		showDialog(templateName, "")
	} else {
		showDialog(service.TemplateFreeForm, "")
	}
}

// Helper: show edit dialog with fields from decrypted content
func showEditDialogWithContent(w fyne.Window, entryName, content string, onSave func(values map[string]string)) {
	templateName := getTemplateFromContent(content)
	_ = entryName
	tmpl := service.GetTemplateByName(templateName)
	if tmpl == nil {
		tmpl = service.GetTemplateByName(service.TemplateFreeForm)
	}
	values := parseFieldsFromContent(content, tmpl)
	fieldWidgets := map[string]*widget.Entry{}
	var formItems []*widget.FormItem
	c := cases.Title(language.English)
	// Minimum width is now hardcoded as 400 in entry.Resize
	for _, field := range tmpl.Fields {
		var entry *widget.Entry
		value := values[field]
		// For Free Form or missing metadata, use raw content if value is empty
		if (tmpl.Name == service.TemplateFreeForm || tmpl.Name == "") && field == "content" && value == "" && content != "" {
			value = content
		}
		if field == "content" || field == "extra" {
			entry = widget.NewMultiLineEntry()
			entry.SetText(value)
			entry.Wrapping = fyne.TextWrapWord
			entry.Resize(fyne.NewSize(400, 60))
		} else {
			entry = widget.NewEntry()
			entry.SetText(value)
			entry.Resize(fyne.NewSize(400, 30))
		}
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


func checkPasswordStoreAndInitIfNotExist(a *ui.App) fyne.CanvasObject {
	// Show setup message and Init button
	info := widget.NewLabel("Password store is not set up yet. Please press Init to start.")
	info.Wrapping = fyne.TextWrapWord

	initBtn := widget.NewButton("Init", func() {
		// Show dialog for store creation
		folderEntry := widget.NewEntry()
		folderEntry.SetText(config.PasswordStoreDirName())
		folderEntry.Disable()
		remoteEntry := widget.NewEntry()
		remoteEntry.SetPlaceHolder("git@host:path/to/repo.git")
		keyIdEntry := widget.NewEntry()
		keyIdEntry.SetPlaceHolder("GPG Key ID")
		help := widget.NewLabel("If you don't have a GPG key, create it using 'gpg --full-generate-key', then run 'gpg -K' to find the key id.")
		form := widget.NewForm(
			widget.NewFormItem("Folder Name", folderEntry),
			widget.NewFormItem("Remote Git URL (SSH)", remoteEntry),
			widget.NewFormItem("Key ID", keyIdEntry),
		)
		dialog.ShowCustomConfirm(
			"Init Password Store", "Confirm", "Cancel",
			container.NewVBox(form, help),
			func(ok bool) {
				if !ok {
					return
				}
				home, _ := os.UserHomeDir()
				baseDir := filepath.Join(home, folderEntry.Text)
				remote := remoteEntry.Text
				keyId := keyIdEntry.Text

				if matched, err := regexp.MatchString(`^git@[\w\.-]+:[\w\./-]+\.git$`, remote); !matched || err != nil {
					ui.ShowErrorDialog(a.Window, fmt.Errorf("remote url doesn't follow expected format"))
					return
				}
				err := store.InitPasswordStore(baseDir, keyId, remote)
				if err != nil {
					ui.ShowErrorDialog(a.Window, fmt.Errorf("failed to init password store: %w", err))
					return
				}

				content := container.NewVBox(widget.NewLabel("Password store is created successfully!"))
				d := dialog.NewCustom("Success", "Cancel", content, a.Window)
				d.SetButtons([]fyne.CanvasObject{widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {
					d.Hide()
					a.ShowScreen("Main")
				})})

				d.Show()
			}, a.Window)
	})

	infoWrap := container.NewGridWrap(fyne.NewSize(400, 60), info)

	vbox := container.NewVBox(
		infoWrap,
		initBtn,
	)
	vbox.Resize(fyne.NewSize(500, 200))

	return vbox
}

func mainUI(a *ui.App) fyne.CanvasObject {
	status := widget.NewLabel("")
	entries, _ := listPasswordEntries(config.PasswordStoreDir())
	selectedIdx := -1
	var decryptBtn, addBtn, editBtn, deleteBtn, syncBtn *widget.Button

	entriesList := widget.NewList(
		func() int {
			entries, _ = listPasswordEntries(config.PasswordStoreDir())
			return len(entries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			entries, _ = listPasswordEntries(config.PasswordStoreDir())
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
		if editBtn != nil {
			editBtn.Enable()
		}
		if deleteBtn != nil {
			deleteBtn.Enable()
		}
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		entriesList.Refresh()
		status.SetText("Entries refreshed")
	})

	syncBtn = widget.NewButton("Sync", func() {
		err := git.SyncWithRemote(config.PasswordStoreDir())
		if err != nil {
			dialog.ShowError(err, a.Window)
			return
		}
		status.SetText("Sync completed")
	})

	addBtn = widget.NewButton("Add", func() {
		showAddOrEditDialog(a.Window, "Add Entry", "Add", "Cancel", "", "", nil, func(entryName, templateName string, values map[string]string) {
			req := service.AddEditRequest{
				EntryName:    entryName,
				TemplateName: templateName,
				Fields:       values,
			}
			result := service.AddOrEditEntry(req)
			if result.Err != nil {
				dialog.ShowError(result.Err, a.Window)
				return
			}
			entriesList.Refresh()
			status.SetText("Entry added")
		})
	})

	editBtn = widget.NewButton("Edit", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), a.Window)
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
					result := service.DecryptAndCacheIfOk(entry, pass)
					if result.Err != nil {
						dialog.ShowError(result.Err, a.Window)
						return
					}
					showEditDialogWithContent(a.Window, entry, result.Plaintext, func(values map[string]string) {
						req := service.AddEditRequest{
							EntryName:    entry,
							TemplateName: getTemplateFromContent(result.Plaintext),
							Fields:       values,
						}
						editResult := service.AddOrEditEntry(req)
						if editResult.Err != nil {
							dialog.ShowError(editResult.Err, a.Window)
							return
						}
						entriesList.Refresh()
						status.SetText("Entry updated")
					})
				}, a.Window)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
		result := service.DecryptAndCacheIfOk(entry, pass)
		if result.Err != nil {
			dialog.ShowError(result.Err, a.Window)
			return
		}
		showEditDialogWithContent(a.Window, entry, result.Plaintext, func(values map[string]string) {
			req := service.AddEditRequest{
				EntryName:    entry,
				TemplateName: getTemplateFromContent(result.Plaintext),
				Fields:       values,
			}
			editResult := service.AddOrEditEntry(req)
			if editResult.Err != nil {
				dialog.ShowError(editResult.Err, a.Window)
				return
			}
			entriesList.Refresh()
			status.SetText("Entry updated")
		})
	})
	editBtn.Disable()

	deleteBtn = widget.NewButton("Delete", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), a.Window)
			return
		}
		entry := entries[selectedIdx]
		dialog.ShowConfirm("Delete Entry", fmt.Sprintf("Are you sure you want to delete '%s'?", entry), func(ok bool) {
			if !ok {
				return
			}
			err := service.DeleteEntry(entry)
			if err != nil {
				dialog.ShowError(err, a.Window)
				return
			}
			entriesList.Refresh()
			status.SetText("Entry deleted")
		}, a.Window)
	})
	deleteBtn.Disable()

	showDecryptedDialog := func(parent fyne.Window, text string) {
		templateName := getTemplateFromContent(text)
		tmpl := service.GetTemplateByName(templateName)
		if tmpl == nil {
			tmpl = service.GetTemplateByName(service.TemplateFreeForm)
		}
		values := parseFieldsFromContent(text, tmpl)
		var items []fyne.CanvasObject
		c := cases.Title(language.English)
		clipboard := parent.Clipboard()
		// Minimum width is now hardcoded as 400 in entry.Resize
		for _, field := range tmpl.Fields {
			value := values[field]
			// For Free Form or missing metadata, use raw text if value is empty
			if (tmpl.Name == service.TemplateFreeForm || tmpl.Name == "") && field == "content" && value == "" && text != "" {
				value = text
			}
			label := widget.NewLabel(fmt.Sprintf("%s:", c.String(field)))
			if field == "content" || field == "extra" {
				entry := widget.NewMultiLineEntry()
				entry.SetText(value)
				entry.Wrapping = fyne.TextWrapWord
				entryWrap := container.NewGridWrap(fyne.NewSize(400, 60), entry)
				copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					clipboard.SetContent(entry.Text)
				})
				row := container.NewHBox(entryWrap, copyBtn)
				items = append(items, label, row)
			} else {
				entry := widget.NewEntry()
				entry.SetText(value)
				entryWrap := container.NewGridWrap(fyne.NewSize(400, 40), entry)
				copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					clipboard.SetContent(entry.Text)
				})
				row := container.NewHBox(entryWrap, copyBtn)
				items = append(items, label, row)
			}
		}
		content := container.NewVBox(items...)
		d := dialog.NewCustom("Decrypted", "OK", content, parent)
		d.Resize(fyne.NewSize(500, 400))
		d.Show()
	}

	decryptBtn = widget.NewButton("Decrypt", func() {
		if selectedIdx < 0 || selectedIdx >= len(entries) {
			dialog.ShowError(fmt.Errorf("no entry selected"), a.Window)
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
					// Only cache if successful
					result := service.DecryptAndCacheIfOk(entry, pass)
					if result.Err != nil {
						dialog.ShowError(result.Err, a.Window)
						return
					}
					showDecryptedDialog(a.Window, result.Plaintext)
				}, a.Window)
			d.Resize(fyne.NewSize(400, 200))
			d.Show()
			return
		}
		result := service.DecryptAndCacheIfOk(entry, pass)
		if result.Err != nil {
			dialog.ShowError(result.Err, a.Window)
			return
		}
		showDecryptedDialog(a.Window, result.Plaintext)
	})
	decryptBtn.Disable()

	// Add the buttons to the UI so they are used
	btnRow := container.NewHBox(refreshBtn, addBtn, editBtn, deleteBtn, decryptBtn, syncBtn)
	entriesLabel := widget.NewLabel("Password Entries")
	entriesScroll := container.NewVScroll(entriesList)
	entriesScroll.SetMinSize(fyne.NewSize(400, 300))
	mainContent := container.NewVBox(
		btnRow,
		entriesLabel,
		entriesScroll,
	)
	content := container.NewBorder(nil, status, nil, nil, mainContent)

	// Add File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Quit", func() { a.Window.Close() }),
	)
	gitMenu := fyne.NewMenu("Git",
		fyne.NewMenuItem("Sync", func() {
			err := git.SyncWithRemote(config.PasswordStoreDir())
			if err != nil {
				dialog.ShowError(err, a.Window)
				return
			}
			status.SetText("Sync completed")
		}),
	)
	mainMenu := fyne.NewMainMenu(fileMenu, gitMenu)
	a.Window.SetMainMenu(mainMenu)

	return content
}

func main() {
	a := app.New()
	w := a.NewWindow("GoPass UI MVP")
	screens := ui.NewScreens()
	screens.AddScreen("Main", mainUI)
	screens.AddScreen("InitStore", checkPasswordStoreAndInitIfNotExist)

	app := &ui.App{Window: w, Screens: screens}

	if !gpg.CheckGPGAvailable() {
		ui.ShowErrorDialog(w, fmt.Errorf("gpg command not found. Please install GnuPG (gpg) and restart the app"))
		return
	}

	_, err := os.Stat(config.PasswordStoreDir())
	slog.Info("Checking password store directory", "path", config.PasswordStoreDir(), "error", err)

	if err != nil {
		// If error is something else than not exist, show error and exit
		if os.IsNotExist(err) {
			slog.Info("Password store directory does not exist, showing Init screen")
			app.ShowScreen("InitStore")
		} else {
			ui.ShowErrorDialog(app.Window, fmt.Errorf("failed to access password store directory: %w", err))
			return
		}
	} else {
		app.ShowScreen("Main")
	}

	w.ShowAndRun()
}
