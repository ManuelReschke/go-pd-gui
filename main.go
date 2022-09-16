package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ManuelReschke/go-pd/pkg/pd"
	"image/color"
	"net/url"
	"strings"
	"time"
)

const (
	Version        = "v1.1.0"
	AppID          = "com.gopdgui.app"
	WindowTitle    = "Go-PD-GUI - DRAINY - PixelDrain Upload Tool"
	WindowWidth    = 550
	WindowHeight   = 600
	SettingAPIKey  = "setting.apikey"
	Headline       = "PixelDrain.com Upload Tool"
	FormLabel      = "API KEY:"
	FormLabelInput = "*optional"
	ButtonCopy     = "Copy"
	ButtonUpload   = "Upload"
	EmptyString    = ""
	FooterText     = "This tool was made by Manuel Reschke under MIT Licence. "
	AboutText      = "@Author: Manuel Reschke\n " +
		"@Github: https://github.com/ManuelReschke/go-pd-gui\n\n " +
		"This tool was made by Manuel Reschke under MIT Licence.\n\n " +
		"Version: " + Version
)

type Settings struct {
	APIKey binding.String
}

type UploadHistory struct {
	UploadDate time.Time
	FileName   string
	URL        string
}

type Storage struct {
	LastElement   UploadHistory
	UploadHistory []UploadHistory
}

type AppData struct {
	App      fyne.App
	Window   fyne.Window
	Settings Settings
	Storage  Storage
}

var MyApp AppData

func main() {
	MyApp.App = app.NewWithID(AppID)
	MyApp.App.Settings().SetTheme(theme.DarkTheme())

	MyApp.Window = MyApp.App.NewWindow(WindowTitle)
	MyApp.Window.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	MyApp.Settings.APIKey = binding.NewString()
	fmt.Println(MyApp.Settings.APIKey)
	_ = MyApp.Settings.APIKey.Set("TEST")
	fmt.Println(MyApp.Settings.APIKey)
	a, _ := MyApp.Settings.APIKey.Get()
	fmt.Println(a)

	// Main Menu
	MyApp.Settings.APIKey = binding.NewString()
	menuItemSettings := fyne.NewMenuItem("Settings", menuActionSettings)
	menuItemAbout := fyne.NewMenuItem("About", menuActionAbout)

	test := fyne.NewMenu("> Menu", menuItemSettings, menuItemAbout)
	mainMenu := fyne.NewMainMenu(test)
	MyApp.Window.SetMainMenu(mainMenu)

	// binding result link
	linkBound := binding.NewString()

	copySuccessWidget := widget.NewIcon(theme.ConfirmIcon())
	containerCopySuccess := container.New(layout.NewCenterLayout(), copySuccessWidget)
	containerCopySuccess.Hide()

	// LINK RESULT
	copyLinkButton := widget.NewButtonWithIcon(ButtonCopy, theme.ContentCopyIcon(), func() {
		if content, err := linkBound.Get(); err == nil {
			MyApp.Window.Clipboard().SetContent(content)
			containerCopySuccess.Show()
			go func() {
				if containerCopySuccess.Visible() {
					time.Sleep(3 * time.Second)
					containerCopySuccess.Hide()
				}
			}()
		}
	})
	containerLinkText := container.New(layout.NewCenterLayout(), copyLinkButton)

	resultWidget := widget.NewHyperlink(EmptyString, nil)
	containerLinkBox := container.New(layout.NewCenterLayout(), resultWidget)

	resultContainer := container.New(layout.NewHBoxLayout(), containerLinkBox, containerLinkText, containerCopySuccess)
	resultContainer.Hide()

	// PROGRESSBAR
	progressBar := widget.NewProgressBarInfinite()
	containerProgressBar := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), progressBar, layout.NewSpacer())
	containerProgressBar.Hide() // hide per default

	// UPLOAD BUTTON ACTION
	uploadButton := widget.NewButtonWithIcon(ButtonUpload, theme.UploadIcon(), func() {
		fileOpen := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if closer != nil {
				defer closer.Close()
				containerProgressBar.Show()

				key, _ := MyApp.Settings.APIKey.Get()
				//cleanKey := strings.TrimSpace(key)
				//
				//// store user input
				//MyApp.App.Preferences().SetString(SettingAPIKey, cleanKey)

				canRead, err := storage.CanRead(closer.URI())
				if err != nil {
					containerProgressBar.Hide()
					dialog.ShowError(err, MyApp.Window)
					return
				}

				if canRead {
					r, err := storage.Reader(closer.URI())
					if err != nil {
						containerProgressBar.Hide()
						dialog.ShowError(err, MyApp.Window)
						return
					}

					pixelDrainURL, err := upload(r, key)
					if err != nil {
						containerProgressBar.Hide()
						dialog.ShowError(err, MyApp.Window)
						return
					}

					_ = linkBound.Set(pixelDrainURL) // set Link
					parsedURL, _ := url.Parse(pixelDrainURL)
					resultWidget.SetText(pixelDrainURL)
					resultWidget.SetURL(parsedURL)
					resultContainer.Show()
					containerProgressBar.Hide()

					// save last element in storage
					lastElement := UploadHistory{
						UploadDate: time.Now(),
						FileName:   "image",
						URL:        pixelDrainURL,
					}
					MyApp.Storage.LastElement = lastElement
				}
			}
			return
		}, MyApp.Window)
		fileOpen.Resize(fyne.NewSize(600, 600))
		fileOpen.Show()
	})
	containerButton := container.New(layout.NewCenterLayout(), uploadButton)

	// History Container
	data := []string{"test test test test", "test test"}
	historyList := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			w := widget.NewLabel("template")
			return w
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i])
		})

	testButton := widget.NewButtonWithIcon("Show History", theme.InfoIcon(), func() {
		window := MyApp.App.NewWindow("test")
		content := container.New(layout.NewBorderLayout(nil, nil, nil, nil), historyList)
		window.SetContent(content)
		window.Resize(fyne.NewSize(480, 380))
		window.Show()
	})

	uploadHistoryContainer := container.New(layout.NewCenterLayout(), testButton)

	container00 := container.New(
		layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildLogo(), layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildHeader(), layout.NewSpacer()),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildHeader2(), layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerButton, layout.NewSpacer()),
		containerProgressBar,
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), resultContainer, layout.NewSpacer()),
		layout.NewSpacer(),
		uploadHistoryContainer,
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildCopyright(), layout.NewSpacer()),
	)

	MyApp.Window.SetContent(container00)
	MyApp.Window.ShowAndRun()
}

func buildLogo() *fyne.Container {
	image := canvas.NewImageFromResource(resourceGoPdGuiIconPng)
	image.SetMinSize(fyne.NewSize(256, 256))
	image.FillMode = canvas.ImageFillStretch
	containerLogo := container.New(layout.NewCenterLayout(), image)

	return containerLogo
}

func buildHeader() *fyne.Container {
	headline := canvas.NewText(Headline, color.White)
	return container.New(layout.NewCenterLayout(), headline)
}

func buildHeader2() *fyne.Container {
	headline2 := canvas.NewText("Hello Anonym, lets upload some files!", color.White)
	return container.New(layout.NewCenterLayout(), headline2)
}

func buildCopyright() *fyne.Container {
	colorLightSilver := color.RGBA{R: 79, G: 79, B: 79, A: 0}
	copyright := canvas.NewText(FooterText+Version, colorLightSilver)
	copyright.TextSize = 11
	containerEnd := container.New(layout.NewCenterLayout(), copyright)
	containerEnd.Resize(fyne.NewSize(0, 50))

	return containerEnd
}

func upload(urc fyne.URIReadCloser, key string) (string, error) {
	fileName := urc.URI().Name()
	if fileName == "" {
		return "", errors.New("filename can not be empty")
	}

	req := &pd.RequestUpload{
		File:      urc,
		FileName:  fileName,
		Anonymous: true,
	}
	if key != "" {
		req.Anonymous = false
		req.Auth.APIKey = key
	}

	c := pd.New(nil, nil)
	rsp, err := c.UploadPOST(req)
	if err != nil {
		return "", err
	}

	return rsp.GetFileURL(), nil
}

// handle the settings window
func menuActionSettings() {
	w := MyApp.App.NewWindow("Settings")

	// FORM INPUT API-KEY
	formLabelAPIKey := canvas.NewText(FormLabel, color.White)

	inputAPIey := widget.NewEntry()
	inputAPIey.SetPlaceHolder(FormLabelInput)
	//MyApp.Settings.APIKey = binding.NewString()
	_ = MyApp.Settings.APIKey.Set(MyApp.App.Preferences().StringWithFallback(SettingAPIKey, ""))
	inputAPIey.Bind(MyApp.Settings.APIKey)

	containerForm := container.New(layout.NewFormLayout(), formLabelAPIKey, inputAPIey)
	containerForm.Resize(fyne.NewSize(0, 200))

	closeButton := widget.NewButtonWithIcon("Save", theme.ContentClearIcon(), func() {
		key, _ := MyApp.Settings.APIKey.Get()
		MyApp.App.Preferences().SetString(SettingAPIKey, strings.TrimSpace(key)) // store user input
		w.Hide()
	})
	containerButton := container.New(layout.NewCenterLayout(), closeButton)

	content := container.New(layout.NewVBoxLayout(), containerForm, layout.NewSpacer(), containerButton)

	w.SetContent(content)
	w.Resize(fyne.NewSize(480, 380))
	w.Show()
}

func menuActionAbout() {
	dialog.ShowInformation("About", AboutText, MyApp.Window)
}
