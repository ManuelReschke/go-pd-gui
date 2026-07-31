package gui

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ManuelReschke/go-pd-gui/internal/fileprep"
	"github.com/ManuelReschke/go-pd/pkg/pd"
	req2 "github.com/imroc/req/v3"
)

const (
	Version      = "v1.3.0"
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
	loadAndApplyTheme()

	MyApp.Window = MyApp.App.NewWindow(WindowTitle)
	MyApp.Window.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	// BINDINGS
	MyApp.Settings.APIKey = binding.NewString()
	MyApp.Settings.Username = binding.NewString()
	MyApp.Settings.Subscription = binding.NewString()
	// Seed from preferences so upload works without opening Settings first.
	_ = MyApp.Settings.APIKey.Set(MyApp.App.Preferences().StringWithFallback(SettingAPIKey, ""))

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
				time.Sleep(3 * time.Second)
				fyne.Do(func() {
					if containerCopySuccess.Visible() {
						containerCopySuccess.Hide()
					}
				})
			}()
		}
	})
	containerLinkText := container.New(layout.NewCenterLayout(), copyLinkButton)

	resultWidget := widget.NewHyperlink(EmptyString, nil)
	containerLinkBox := container.New(layout.NewCenterLayout(), resultWidget)

	resultContainer := container.New(layout.NewHBoxLayout(), containerLinkBox, containerLinkText, containerCopySuccess)
	resultContainer.Hide()

	// PROGRESSBAR
	progressBar := widget.NewProgressBar()
	containerProgressBar := container.New(layout.NewVBoxLayout(), layout.NewSpacer(), progressBar, layout.NewSpacer())
	containerProgressBar.Hide() // hide per default

	// LAST ELEMENT
	lastElementContainer := container.New(layout.NewHBoxLayout())
	lastElementContainer.Hide()

	// UPLOAD BUTTON ACTION
	// Upload must run off the UI thread: the file dialog only closes after this
	// callback returns. Blocking here freezes the picker until the transfer ends.
	// Declare first so the callback can call Disable/Enable (:= scope starts after the statement).
	var uploadButton *widget.Button
	uploadButton = widget.NewButtonWithIcon(ButtonUpload, theme.UploadIcon(), func() {
		fileOpen := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, MyApp.Window)
				return
			}
			if reader == nil {
				return // user cancelled
			}

			key, _ := MyApp.Settings.APIKey.Get()
			if key == "" {
				// Binding may be empty until Settings is opened; fall back to prefs.
				key = MyApp.App.Preferences().String(SettingAPIKey)
			}

			progressBar.SetValue(0)
			containerProgressBar.Show()
			uploadButton.Disable()

			// Materialize + upload off the UI thread. On Android, SAF content URIs are
			// not real filesystem paths; materializeForUpload copies them to a temp file
			// so go-pd can use PathToFile / SetFile.
			go func() {
				defer fyne.Do(func() {
					uploadButton.Enable()
					containerProgressBar.Hide()
				})

				filePath, fileName, cleanup, matErr := materializeForUpload(reader)
				if matErr != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("cannot prepare file for upload: %w", matErr), MyApp.Window)
					})
					return
				}
				defer cleanup()

				pixelDrainURL, err := upload(filePath, fileName, key, progressBar)
				if err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("upload failed: %w", err), MyApp.Window)
					})
					return
				}

				fyne.Do(func() {
					_ = linkBound.Set(pixelDrainURL)
					parsedURL, _ := url.Parse(pixelDrainURL)
					resultWidget.SetText(pixelDrainURL)
					resultWidget.SetURL(parsedURL)
					resultContainer.Show()

					lastElement := UploadHistory{
						UploadDate: time.Now(),
						FileName:   fileName,
						URL:        pixelDrainURL,
					}
					MyApp.Storage.LastElement = lastElement

					lastElementContainer.Objects = nil
					lastElementURLParsed, _ := url.Parse(pixelDrainURL)
					lastElementText := widget.NewLabel("Last upload:")
					lastElementLink := widget.NewHyperlink(lastElement.FileName, lastElementURLParsed)
					lastElementContainer.Add(lastElementText)
					lastElementContainer.Add(lastElementLink)
					lastElementContainer.Show()
					lastElementContainer.Refresh()
				})
			}()
		}, MyApp.Window)
		// In Fyne v2.8+, Resize before Show panics: MinSize() needs a built dialog.
		// Show creates the dialog; Resize afterward applies the preferred size.
		fileOpen.Show()
		fileOpen.Resize(fyne.NewSize(600, 600))
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
	headline := widget.NewLabel(Headline)
	headline.TextStyle = fyne.TextStyle{Bold: true}
	return container.New(layout.NewCenterLayout(), headline)
}

func buildHeader2() *fyne.Container {
	usern := MyApp.App.Preferences().String(SettingUsername)
	if usern != "" {
		_ = MyApp.Settings.Username.Set(usern)
	}

	textStart := widget.NewLabel("Hello")
	textEnd := widget.NewLabel(", lets upload some files!")
	usernameBinding := widget.NewLabelWithData(MyApp.Settings.Username)
	return container.New(layout.NewHBoxLayout(), textStart, usernameBinding, textEnd)
}

func buildCopyright() *fyne.Container {
	copyright := canvas.NewText(FooterText+Version, theme.DisabledColor())
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

// materializeForUpload returns a real local filesystem path for go-pd's
// PathToFile / SetFile upload path.
//
// Desktop file:// URIs already map to readable paths. On Android (and other SAF
// providers) the dialog yields content://document URIs such as
// /document/primary%3APictures%2Fphoto.png which cannot be opened with os.Open.
// In that case the URI reader is copied into a temporary file that the caller
// must remove via the returned cleanup function.
//
// The reader is always closed by this function.
func materializeForUpload(reader fyne.URIReadCloser) (localPath, fileName string, cleanup func(), err error) {
	cleanup = func() {}
	if reader == nil {
		return "", "", cleanup, errors.New("file reader is nil")
	}
	defer func() {
		// Ensure the picker stream is closed even if we already consumed it.
		_ = reader.Close()
	}()

	fileURI := reader.URI()
	fileName = fileprep.SanitizeFileName(fileURI.Name())

	if path, ok := fileprep.ReadableLocalFile(fileURI.Path()); ok {
		return path, fileName, cleanup, nil
	}

	// Android SAF / content URIs: stream into a temp file for go-pd SetFile.
	tmpPath, copyErr := fileprep.CopyReaderToTempFile(reader, fileName)
	if copyErr != nil {
		return "", "", cleanup, fmt.Errorf("copy from %q: %w", fileURI.String(), copyErr)
	}
	cleanup = func() { _ = os.Remove(tmpPath) }
	return tmpPath, fileName, cleanup, nil
}

// upload sends a local file to PixelDrain.
//
// go-pd UploadPOST always uses PathToFile with req.SetFile (the File reader is
// not used for the HTTP body). That is required for correct progress FileSize.
func upload(filePath, fileName, key string, progressBar *widget.ProgressBar) (string, error) {
	if filePath == "" {
		return "", errors.New("file path can not be empty")
	}
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	if fileName == "" || fileName == "." || fileName == "/" {
		return "", errors.New("filename can not be empty")
	}

	req := &pd.RequestUpload{
		PathToFile: filePath,
		FileName:   fileName,
		Anonymous:  true,
	}
	if key != "" {
		req.Anonymous = false
		req.Auth.APIKey = key
	}

	c := pd.New(nil, nil)
	c.SetUploadCallback(func(info req2.UploadInfo) {
		if info.FileSize <= 0 {
			return
		}
		fyne.Do(func() {
			progressBar.SetValue(float64(info.UploadedSize) / float64(info.FileSize))
		})
	})
	rsp, err := c.UploadPOST(req)
	if err != nil {
		return "", err
	}
	if rsp == nil {
		return "", errors.New("empty response from pixeldrain")
	}

	return rsp.GetFileURL(), nil
}
