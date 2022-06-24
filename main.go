package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"image/color"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Grid Layout")
	myWindow.Resize(fyne.NewSize(550, 600))

	image := canvas.NewImageFromFile("go-pd-gui-logo.png")
	image.SetMinSize(fyne.NewSize(400, 200))
	image.FillMode = canvas.ImageFillOriginal
	container01 := container.New(layout.NewCenterLayout(), image)

	formLabelAPIKey := canvas.NewText("PixelDrain API KEY:", color.Black)
	formAPIey := canvas.NewText("Value", color.White)
	container02 := container.New(layout.NewFormLayout(), formLabelAPIKey, formAPIey)
	container02.Resize(fyne.NewSize(0, 200))

	copyright := canvas.NewText("This tool was made by Manuel Reschke under MIT Licence", color.White)
	containerEnd := container.New(layout.NewCenterLayout(), copyright)
	containerEnd.Resize(fyne.NewSize(0, 50))

	container00 := container.New(layout.NewGridLayout(1), container01, container02, containerEnd)
	myWindow.SetContent(container00)
	myWindow.ShowAndRun()
	//f := app.New()
	//w := f.NewWindow("")
	//label1 := widget.NewLabel("Label1")
	//
	//b1 := widget.NewButton("Button1", func() {})
	//b2 := widget.NewButton("Button2", func() {})
	//label2 := widget.NewLabel("Label3")
	//
	//w.SetContent(
	//	fyne.NewContainerWithLayout(
	//		layout.NewVBoxLayout(),
	//		fyne.NewContainerWithLayout(layout.NewHBoxLayout(), layout.NewSpacer(), label1, layout.NewSpacer()),
	//		layout.NewSpacer(),
	//		fyne.NewContainerWithLayout(layout.NewHBoxLayout(), layout.NewSpacer(), b1, b2, layout.NewSpacer()),
	//		layout.NewSpacer(),
	//		fyne.NewContainerWithLayout(layout.NewHBoxLayout(), layout.NewSpacer(), label2, layout.NewSpacer()),
	//	),
	//)
	//
	//w.Resize(fyne.Size{Height: 320, Width: 480})
	//
	//w.ShowAndRun()
}
