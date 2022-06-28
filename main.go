package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ManuelReschke/go-pd/pkg/pd"
	"image/color"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	Version      = "v0.1.0"
	AppID        = "com.gopdgui.app"
	WindowTitle  = "Go-PD-GUI - PixelDrain Upload Tool"
	WindowWidth  = 550
	WindowHeight = 600

	SettingAPIKey = "setting.apikey"

	Headline = "PixelDrain.com Upload Tool"

	FormLabel      = "API KEY:"
	FormLabelInput = "*optional"

	ButtonCopy   = "Copy"
	ButtonUpload = "Upload"

	EmptyString = ""

	AssetLogo = "assets/go-pd-gui-logo.png"

	FooterText = "This tool was made by Manuel Reschke under MIT Licence. "
)

func main() {
	myApp := app.NewWithID(AppID)
	myWindow := myApp.NewWindow(WindowTitle)
	myWindow.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	// FORM INPUT API-KEY
	formLabelAPIKey := canvas.NewText(FormLabel, color.White)
	inputAPIey := widget.NewEntry()
	inputAPIey.SetPlaceHolder(FormLabelInput)
	apiKeyBinding := binding.NewString()
	_ = apiKeyBinding.Set(myApp.Preferences().StringWithFallback(SettingAPIKey, ""))
	inputAPIey.Bind(apiKeyBinding)
	containerForm := container.New(layout.NewFormLayout(), formLabelAPIKey, inputAPIey)
	containerForm.Resize(fyne.NewSize(0, 200))

	// binding result link
	linkBound := binding.NewString()

	copySuccessWidget := widget.NewIcon(theme.ConfirmIcon())
	containerCopySuccess := container.New(layout.NewCenterLayout(), copySuccessWidget)
	containerCopySuccess.Hide()

	// LINK RESULT
	copyLinkButton := widget.NewButtonWithIcon(ButtonCopy, theme.ContentCopyIcon(), func() {
		if content, err := linkBound.Get(); err == nil {
			myWindow.Clipboard().SetContent(content)
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

				key, _ := apiKeyBinding.Get()
				cleanKey := strings.TrimSpace(key)

				// store user input
				myApp.Preferences().SetString(SettingAPIKey, cleanKey)

				pixelDrainURL, err := upload(closer, cleanKey)
				if err != nil {
					containerProgressBar.Hide()
					dialog.ShowError(err, myWindow)
					return
				}

				_ = linkBound.Set(pixelDrainURL) // set Link
				parsedURL, _ := url.Parse(pixelDrainURL)
				resultWidget.SetText(pixelDrainURL)
				resultWidget.SetURL(parsedURL)
				resultContainer.Show()
				containerProgressBar.Hide()
			}
			return
		}, myWindow)
		fileOpen.Resize(fyne.NewSize(600, 600))
		fileOpen.Show()
	})
	containerButton := container.New(layout.NewCenterLayout(), uploadButton)

	container00 := container.New(
		layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildLogo(), layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildHeader(), layout.NewSpacer()),
		container.New(layout.NewVBoxLayout(), layout.NewSpacer(), containerForm, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerButton, layout.NewSpacer()),
		containerProgressBar,
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), resultContainer, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), buildCopyright(), layout.NewSpacer()),
	)

	myWindow.SetContent(container00)
	myWindow.ShowAndRun()
}

func buildLogo() *fyne.Container {
	image := canvas.NewImageFromFile(filepath.FromSlash(AssetLogo))
	image.SetMinSize(fyne.NewSize(361, 152))
	image.FillMode = canvas.ImageFillStretch
	containerLogo := container.New(layout.NewCenterLayout(), image)

	return containerLogo
}

func buildHeader() *fyne.Container {
	headline := canvas.NewText(Headline, color.White)
	return container.New(layout.NewCenterLayout(), headline)
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
