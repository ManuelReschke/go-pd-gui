package main

import (
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
	"time"
)

const VERSION = "v0.1.0"

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go-PD-GUI - PixelDrain Upload Tool")
	myWindow.Resize(fyne.NewSize(550, 600))

	headline := canvas.NewText("PixelDrain.com Upload Tool", color.White)
	containerHeading := container.New(layout.NewCenterLayout(), headline)

	formLabelAPIKey := canvas.NewText("API KEY:", color.White)
	inputAPIey := widget.NewEntry()
	inputAPIey.SetPlaceHolder("Enter text...")
	containerForm := container.New(layout.NewFormLayout(), formLabelAPIKey, inputAPIey)
	containerForm.Resize(fyne.NewSize(0, 200))

	// binding result link
	linkBound := binding.NewString()

	copySuccessWidget := widget.NewIcon(theme.ConfirmIcon())
	containerCopySuccess := container.New(layout.NewCenterLayout(), copySuccessWidget)
	containerCopySuccess.Hide()

	// LINK RESULT
	copyLinkButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
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

	resultWidget := widget.NewHyperlink("", nil)
	containerLinkBox := container.New(layout.NewCenterLayout(), resultWidget)

	resultContainer := container.New(layout.NewHBoxLayout(), containerLinkBox, containerLinkText, containerCopySuccess)
	resultContainer.Hide()

	// PROGRESSBAR
	progressBar := widget.NewProgressBarInfinite()
	containerProgressBar := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), progressBar, layout.NewSpacer())
	containerProgressBar.Hide() // hide per default

	// UPLOAD BUTTON ACTION
	uploadButton := widget.NewButtonWithIcon("Upload", theme.UploadIcon(), func() {
		pixelDrainURL := ""
		fileOpen := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
			if closer != nil {
				defer closer.Close()
				containerProgressBar.Show()

				req := &pd.RequestUpload{
					PathToFile: closer.URI().Path(),
					Anonymous:  true,
				}

				c := pd.New(nil, nil)
				rsp, err := c.UploadPOST(req)
				if err != nil {
					containerProgressBar.Hide()
					dialog.ShowError(err, myWindow)
					return
				}

				pixelDrainURL = rsp.GetFileURL()

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
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerHeading, layout.NewSpacer()),
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
	image := canvas.NewImageFromFile(filepath.FromSlash("assets/go-pd-gui-logo.png"))
	image.SetMinSize(fyne.NewSize(361, 152))
	image.FillMode = canvas.ImageFillStretch
	containerLogo := container.New(layout.NewCenterLayout(), image)

	return containerLogo
}

func buildCopyright() *fyne.Container {
	colorLightSilver := color.RGBA{R: 79, G: 79, B: 79, A: 0}
	copyright := canvas.NewText("This tool was made by Manuel Reschke under MIT Licence. "+VERSION, colorLightSilver)
	copyright.TextSize = 11
	containerEnd := container.New(layout.NewCenterLayout(), copyright)
	containerEnd.Resize(fyne.NewSize(0, 50))

	return containerEnd
}

func upload() (string, error) {
	//err := errors.New("test error")
	return "https://pixeldrain.com/u/YqiUjXBc", nil
}
