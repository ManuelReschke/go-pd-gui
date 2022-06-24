package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"net/url"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go-PD-GUI - PixelDrain Upload Tool")
	myWindow.Resize(fyne.NewSize(550, 600))

	image := canvas.NewImageFromFile("go-pd-gui-logo.png")
	image.SetMinSize(fyne.NewSize(400, 200))
	image.FillMode = canvas.ImageFillOriginal
	containerLogo := container.New(layout.NewCenterLayout(), image)

	headline := canvas.NewText("PixelDrain.com Upload Tool", color.White)
	containerHeading := container.New(layout.NewCenterLayout(), headline)

	formLabelAPIKey := canvas.NewText("API KEY:", color.White)
	inputAPIey := widget.NewEntry()
	inputAPIey.SetPlaceHolder("Enter text...")
	containerForm := container.New(layout.NewFormLayout(), formLabelAPIKey, inputAPIey)
	containerForm.Resize(fyne.NewSize(0, 200))

	resulText := widget.NewLabel("TEST TEST")
	containerLinkText := container.New(layout.NewCenterLayout(), resulText)
	resultWidget := widget.NewHyperlink("", nil)
	containerLinkBox := container.New(layout.NewCenterLayout(), resultWidget)

	resultContainer := container.New(layout.NewHBoxLayout(), containerLinkBox, containerLinkText)
	resultContainer.Hide()

	// PROGRESSBAR
	progressBar := widget.NewProgressBarInfinite()
	containerProgressBar := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), progressBar, layout.NewSpacer())
	containerProgressBar.Hide() // hide per default

	// UPLOAD BUTTON ACTION
	uploadButton := widget.NewButtonWithIcon("Upload", theme.UploadIcon(), func() {
		containerProgressBar.Show()
		pixelDrainURL := "https://pixeldrain.com/u/YqiUjXBc"
		parsedURL, _ := url.Parse(pixelDrainURL)
		resultWidget.SetText(pixelDrainURL)
		resultWidget.SetURL(parsedURL)
		resultContainer.Show()
	})
	containerButton := container.New(layout.NewCenterLayout(), uploadButton)

	copyright := canvas.NewText("This tool was made by Manuel Reschke under MIT Licence", color.RGBA{
		R: 128,
		G: 126,
		B: 126,
		A: 0,
	})
	containerEnd := container.New(layout.NewCenterLayout(), copyright)
	containerEnd.Resize(fyne.NewSize(0, 50))

	container00 := container.New(
		layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerLogo, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerHeading, layout.NewSpacer()),
		container.New(layout.NewVBoxLayout(), layout.NewSpacer(), containerForm, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerButton, layout.NewSpacer()),
		containerProgressBar,
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), resultContainer, layout.NewSpacer()),
		layout.NewSpacer(),
		container.New(layout.NewHBoxLayout(), layout.NewSpacer(), containerEnd, layout.NewSpacer()),
	)

	myWindow.SetContent(container00)
	myWindow.ShowAndRun()
}
