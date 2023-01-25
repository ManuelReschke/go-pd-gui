package gui

import (
	"errors"
	"image/color"
	"net/url"
	"time"

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
)

const (
	Version      = "v1.1.0"
	AppID        = "com.gopdgui.app"
	WindowTitle  = "Go-PD-GUI - DRAINY - PixelDrain Upload Tool"
	WindowWidth  = 550
	WindowHeight = 600

	Headline     = "PixelDrain.com Upload Tool"
	ButtonCopy   = "Copy"
	ButtonUpload = "Upload"
	EmptyString  = ""
	FooterText   = "This tool was made by Manuel Reschke under MIT Licence. "
)

type Settings struct {
	APIKey       binding.String
	Username     binding.String
	Subscription binding.String
}

type UploadHistory struct {
	UploadDate time.Time
	FileName   string
	URL        string
}

type Storage struct {
	Username      string
	LastElement   UploadHistory
	UploadHistory []UploadHistory
}

type AppData struct {
	App        fyne.App
	Window     fyne.Window
	Containers map[string]*fyne.Container
	Settings   Settings
	Storage    Storage
}

var MyApp AppData

func BuildStart() {
	MyApp.App = app.NewWithID(AppID)
	MyApp.App.Settings().SetTheme(theme.DarkTheme())

	MyApp.Window = MyApp.App.NewWindow(WindowTitle)
	MyApp.Window.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	// BINDINGS
	MyApp.Settings.APIKey = binding.NewString()
	MyApp.Settings.Username = binding.NewString()
	MyApp.Settings.Subscription = binding.NewString()

	// Main Menu
	MyApp.Window.SetMainMenu(BuildMainMenu())

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

	// LAST ELEMENT
	lastElementContainer := container.New(layout.NewHBoxLayout())
	lastElementContainer.Hide()

	// UPLOAD BUTTON ACTION
	uploadButton := widget.NewButtonWithIcon(ButtonUpload, theme.UploadIcon(), func() {
		fileOpen := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if closer != nil {
				defer closer.Close()
				containerProgressBar.Show()

				key, _ := MyApp.Settings.APIKey.Get()
				// cleanKey := strings.TrimSpace(key)
				//
				// // store user input
				// MyApp.App.Preferences().SetString(SettingAPIKey, cleanKey)

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
						FileName:   r.URI().Name(),
						URL:        pixelDrainURL,
					}
					MyApp.Storage.LastElement = lastElement

					// show last element
					lastElementURLParsed, _ := url.Parse(pixelDrainURL)
					lastElementText := canvas.NewText("Last upload:", color.White)
					lastElementLink := widget.NewHyperlink(MyApp.Storage.LastElement.FileName, lastElementURLParsed)
					lastElementContainer.Add(lastElementText)
					lastElementContainer.Add(lastElementLink)
					lastElementContainer.Show()
				}
			}
			return
		}, MyApp.Window)
		fileOpen.Resize(fyne.NewSize(600, 600))
		fileOpen.Show()
	})
	containerButton := container.New(layout.NewCenterLayout(), uploadButton)

	// set the container to the app to update text later with refresh
	header2Container := buildHeader2()
	MyApp.Containers = map[string]*fyne.Container{"header2": header2Container}

	container00 := container.New(
		layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildLogo(), layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildHeader(), layout.NewSpacer()),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), header2Container, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerButton, layout.NewSpacer()),
		containerProgressBar,
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), resultContainer, layout.NewSpacer()),
		layout.NewSpacer(),
		buildUploadHistoryContainer(),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), lastElementContainer, layout.NewSpacer()),
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
	usern := MyApp.App.Preferences().String(SettingUsername)
	if usern != "" {
		_ = MyApp.Settings.Username.Set(usern)
	}

	textStart := canvas.NewText("Hello", color.White)
	textEnd := canvas.NewText(", lets upload some files!", color.White)
	usernameBinding := widget.NewLabelWithData(MyApp.Settings.Username)
	return container.New(layout.NewHBoxLayout(), textStart, usernameBinding, textEnd)
}

func buildCopyright() *fyne.Container {
	colorLightSilver := color.RGBA{R: 79, G: 79, B: 79, A: 0}
	copyright := canvas.NewText(FooterText+Version, colorLightSilver)
	copyright.TextSize = 11
	containerEnd := container.New(layout.NewCenterLayout(), copyright)
	containerEnd.Resize(fyne.NewSize(0, 50))

	return containerEnd
}

// History Container
func buildUploadHistoryContainer() *fyne.Container {
	historyButton := widget.NewButtonWithIcon("Show History", theme.InfoIcon(), func() {
		BuildHistoryWindow()
	})

	return container.New(layout.NewCenterLayout(), historyButton)
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
