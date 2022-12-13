package gui

import (
	"errors"
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
	"time"
)

const (
	Version      = "v1.1.0"
	AppID        = "com.gopdgui.app"
	WindowTitle  = "Go-PD-GUI - DRAINY - PixelDrain Upload Tool"
	WindowWidth  = 550
	WindowHeight = 600

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
	App      fyne.App
	Window   fyne.Window
	Settings Settings
	Storage  Storage
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
	//username := MyApp.Settings.Username.
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

func buildUploadHistoryContainer() *fyne.Container {
	// History Container
	data := []string{"12.12.2022 | http://pixeldrain.com/kajsdjaksjd", "12.12.2022 | http://pixeldrain.com/kajsdjaksjd", "12.12.2022 | http://pixeldrain.com/kajsdjaksjd"}
	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i])
		})
	// History Container
	//data := [][]string{
	//	[]string{"#", "Date", "Link"},
	//	[]string{"12.03.2022", "ahsakshdhakshd", "http://pixeldrain.com/kajsdjaksjd"},
	//	//[]string{"12.03.2022", "ahsakshdhakshd", "http://pixeldrain.com/kajsdjaksjd"},
	//}
	//historyList := widget.NewTable(
	//	func() (int, int) {
	//		return len(data), len(data[0])
	//	},
	//	func() fyne.CanvasObject {
	//		c := container.NewCenter()
	//		return c
	//		//w := widget.NewLabel("Template")
	//		//return w
	//	},
	//	func(i widget.TableCellID, o fyne.CanvasObject) {
	//		if i.Col == 0 && i.Row > 0 {
	//			w := widget.NewButton("#", func() {
	//				fmt.Println("click")
	//			})
	//			o.(*fyne.Container).Add(w)
	//			o.(*fyne.Container).Resize(fyne.NewSize(100, 100))
	//			return
	//		}
	//
	//		w := widget.NewLabel(data[i.Row][i.Col])
	//		o.(*fyne.Container).Add(w)
	//		//o.(*fyne.Container).Resize(fyne.NewSize(100, 100))
	//
	//		//o.(*widget.Label).SetText(data[i.Row][i.Col])
	//	})
	//historyList.SetColumnWidth(0, 250)
	//historyList.SetColumnWidth(1, 350)

	testButton := widget.NewButtonWithIcon("Show History", theme.InfoIcon(), func() {
		window := MyApp.App.NewWindow("History")
		//content := container.New(layout.NewCenterLayout(), list)
		window.SetContent(list)
		//window.Resize(fyne.NewSize(480, 380))
		window.Resize(fyne.NewSize(600, 380))
		window.Show()
	})

	return container.New(layout.NewCenterLayout(), testButton)
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