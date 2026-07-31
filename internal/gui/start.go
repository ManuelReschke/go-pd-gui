package gui

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	"github.com/ManuelReschke/go-pd-gui/internal/filepick"
	"github.com/ManuelReschke/go-pd-gui/internal/fileprep"
	"github.com/ManuelReschke/go-pd-gui/internal/uploadutil"
	"github.com/ManuelReschke/go-pd/pkg/pd"
	req2 "github.com/imroc/req/v3"
)

const (
	Version      = "v1.3.0"
	AppID        = "com.gopdgui.app"
	WindowTitle  = "Go-PD-GUI - DRAINY - PixelDrain Upload Tool"
	WindowWidth  = 550
	WindowHeight = 650

	Headline        = "PixelDrain.com Upload Tool"
	ButtonCopy      = "Copy"
	ButtonUpload    = "Upload"
	ButtonAddFiles  = "Add files"
	ButtonUploadAll = "Upload all"
	EmptyString     = ""
	FooterText      = "This tool was made by Manuel Reschke under MIT Licence. "
)

// pendingFile is a local file ready for PixelDrain upload (path may be a temp copy).
type pendingFile struct {
	localPath string
	fileName  string
	cleanup   func()
}

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

	// MULTI-FILE QUEUE
	// Fyne's FileOpen dialog only returns one file per pick, so we accumulate
	// selections in a queue and upload them sequentially on "Upload all".
	var pending []pendingFile
	queueStatus := widget.NewLabel("")
	queueStatus.Hide()
	queueBox := container.NewVBox()
	queueScroll := container.NewVScroll(queueBox)
	queueScroll.SetMinSize(fyne.NewSize(0, 100))
	queueContainer := container.NewVBox(queueStatus, queueScroll)
	queueContainer.Hide()

	// Upload status (e.g. "Uploading 2/5: photo.png")
	uploadStatus := widget.NewLabel("")
	uploadStatus.Alignment = fyne.TextAlignCenter
	uploadStatus.Hide()

	var addFilesButton, uploadAllButton *widget.Button

	// Declare first so the remove-button callback can call refreshQueue recursively
	// (:= would leave the name undefined inside the function body).
	var refreshQueue func()
	refreshQueue = func() {
		queueBox.Objects = nil
		for i := range pending {
			idx := i
			f := pending[idx]
			nameLabel := widget.NewLabel(f.fileName)
			nameLabel.Truncation = fyne.TextTruncateEllipsis
			removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				if idx < 0 || idx >= len(pending) {
					return
				}
				if pending[idx].cleanup != nil {
					pending[idx].cleanup()
				}
				pending = append(pending[:idx], pending[idx+1:]...)
				refreshQueue()
			})
			row := container.NewBorder(nil, nil, nil, removeBtn, nameLabel)
			queueBox.Add(row)
		}
		queueBox.Refresh()

		n := len(pending)
		if n == 0 {
			queueContainer.Hide()
			queueStatus.Hide()
			uploadAllButton.Disable()
			uploadAllButton.SetText(ButtonUploadAll)
			return
		}
		queueStatus.SetText(fmt.Sprintf("%d file(s) selected", n))
		queueStatus.Show()
		queueContainer.Show()
		uploadAllButton.Enable()
		if n == 1 {
			uploadAllButton.SetText(ButtonUpload)
		} else {
			uploadAllButton.SetText(fmt.Sprintf("%s (%d)", ButtonUploadAll, n))
		}
	}

	appendPending := func(filePath, fileName string, cleanup func()) {
		if cleanup == nil {
			cleanup = func() {}
		}
		pending = append(pending, pendingFile{
			localPath: filePath,
			fileName:  fileName,
			cleanup:   cleanup,
		})
		refreshQueue()
	}

	// Fyne single-file picker (Android and any platform without a native multi-picker).
	openFyneSingleFilePicker := func() {
		fileOpen := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, MyApp.Window)
				return
			}
			if reader == nil {
				return // user cancelled
			}

			// Materialize off the UI thread (Android SAF may copy into a temp file).
			go func() {
				filePath, fileName, cleanup, matErr := materializeForUpload(reader)
				if matErr != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("cannot prepare file for upload: %w", matErr), MyApp.Window)
					})
					return
				}

				fyne.Do(func() {
					appendPending(filePath, fileName, cleanup)
				})
			}()
		}, MyApp.Window)
		// In Fyne v2.8+, Resize before Show panics: MinSize() needs a built dialog.
		// Show creates the dialog; Resize afterward applies the preferred size.
		fileOpen.Show()
		fileOpen.Resize(fyne.NewSize(600, 600))
	}

	// Open according to Settings → Filepicker (Fyne default | OS Native Multipicker).
	// Native falls back to Fyne when unsupported (e.g. Android).
	openFilePicker := func() {
		if currentFilePickerMode() == FilePickerFyne {
			openFyneSingleFilePicker()
			return
		}

		go func() {
			paths, err := filepick.OpenMultiple("Select files to upload")
			if errors.Is(err, filepick.ErrUnsupported) {
				fyne.Do(openFyneSingleFilePicker)
				return
			}
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("file picker failed: %w", err), MyApp.Window)
				})
				return
			}
			if len(paths) == 0 {
				return // cancelled
			}

			// Validate paths and stage them on the UI thread.
			type staged struct {
				path, name string
			}
			var okFiles []staged
			var failMsgs []string
			for _, p := range paths {
				local, readable := fileprep.ReadableLocalFile(p)
				if !readable {
					failMsgs = append(failMsgs, fmt.Sprintf("%s: not a readable file", filepath.Base(p)))
					continue
				}
				okFiles = append(okFiles, staged{
					path: local,
					name: fileprep.SanitizeFileName(filepath.Base(local)),
				})
			}

			fyne.Do(func() {
				for _, f := range okFiles {
					appendPending(f.path, f.name, nil)
				}
				if len(failMsgs) > 0 {
					dialog.ShowError(
						fmt.Errorf("%d of %d path(s) could not be added:\n%s",
							len(failMsgs), len(paths), strings.Join(failMsgs, "\n")),
						MyApp.Window,
					)
				}
			})
		}()
	}

	addFilesButton = widget.NewButtonWithIcon(ButtonAddFiles, theme.ContentAddIcon(), func() {
		openFilePicker()
	})

	// Upload must run off the UI thread so the UI stays responsive and the
	// progress bar can update via fyne.Do.
	uploadAllButton = widget.NewButtonWithIcon(ButtonUploadAll, theme.UploadIcon(), func() {
		if len(pending) == 0 {
			return
		}

		key, _ := MyApp.Settings.APIKey.Get()
		if key == "" {
			// Binding may be empty until Settings is opened; fall back to prefs.
			key = MyApp.App.Preferences().String(SettingAPIKey)
		}

		// Take ownership of the queue so the user can stage new files afterwards.
		batch := pending
		pending = nil
		refreshQueue()

		progressBar.SetValue(0)
		containerProgressBar.Show()
		uploadStatus.SetText(fmt.Sprintf("Uploading 0/%d…", len(batch)))
		uploadStatus.Show()
		addFilesButton.Disable()
		uploadAllButton.Disable()

		go func() {
			defer fyne.Do(func() {
				addFilesButton.Enable()
				containerProgressBar.Hide()
				uploadStatus.Hide()
				// Re-apply queue button state (pending may still be empty).
				refreshQueue()
			})

			var (
				successes []UploadHistory
				failMsgs  []string
			)
			total := len(batch)

			for i, file := range batch {
				idx := i
				fyne.Do(func() {
					uploadStatus.SetText(fmt.Sprintf("Uploading %d/%d: %s", idx+1, total, file.fileName))
					// Per-file progress mapped into the overall bar.
					progressBar.SetValue(float64(idx) / float64(total))
				})

				// Wrap the shared progress bar so each file contributes 1/total.
				pixelDrainURL, err := upload(file.localPath, file.fileName, key, progressBar, idx, total)
				if file.cleanup != nil {
					file.cleanup()
				}
				if err != nil {
					failMsgs = append(failMsgs, fmt.Sprintf("%s: %v", file.fileName, err))
					continue
				}

				item := UploadHistory{
					UploadDate: time.Now(),
					FileName:   file.fileName,
					URL:        pixelDrainURL,
				}
				successes = append(successes, item)
			}

			fyne.Do(func() {
				progressBar.SetValue(1)

				if len(successes) > 0 {
					last := successes[len(successes)-1]
					MyApp.Storage.LastElement = last
					MyApp.Storage.UploadHistory = append(MyApp.Storage.UploadHistory, successes...)

					// Result link: last URL. Copy uses all URLs when multi (one per line).
					urls := make([]string, len(successes))
					for i, s := range successes {
						urls[i] = s.URL
					}
					_ = linkBound.Set(uploadutil.JoinURLs(urls))
					parsedURL, _ := url.Parse(last.URL)
					if len(successes) == 1 {
						resultWidget.SetText(last.URL)
					} else {
						resultWidget.SetText(fmt.Sprintf("%d files uploaded (last: %s)", len(successes), last.FileName))
					}
					resultWidget.SetURL(parsedURL)
					resultContainer.Show()

					lastElementContainer.Objects = nil
					lastElementURLParsed, _ := url.Parse(last.URL)
					lastElementText := widget.NewLabel("Last upload:")
					lastElementLink := widget.NewHyperlink(last.FileName, lastElementURLParsed)
					lastElementContainer.Add(lastElementText)
					lastElementContainer.Add(lastElementLink)
					lastElementContainer.Show()
					lastElementContainer.Refresh()
				}

				if len(failMsgs) > 0 {
					msg := fmt.Sprintf("%d of %d upload(s) failed:\n%s", len(failMsgs), total, strings.Join(failMsgs, "\n"))
					dialog.ShowError(errors.New(msg), MyApp.Window)
				}
			})
		}()
	})
	uploadAllButton.Disable()

	containerButton := container.New(layout.NewHBoxLayout(), addFilesButton, uploadAllButton)

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
		queueContainer,
		uploadStatus,
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
//
// fileIndex and fileTotal map this file's progress into a multi-file overall bar:
// overall = (fileIndex + fraction) / fileTotal. Pass fileTotal=1 for a single file.
func upload(filePath, fileName, key string, progressBar *widget.ProgressBar, fileIndex, fileTotal int) (string, error) {
	if filePath == "" {
		return "", errors.New("file path can not be empty")
	}
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	if fileName == "" || fileName == "." || fileName == "/" {
		return "", errors.New("filename can not be empty")
	}
	if fileTotal < 1 {
		fileTotal = 1
	}
	if fileIndex < 0 {
		fileIndex = 0
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
		fraction := float64(info.UploadedSize) / float64(info.FileSize)
		overall := uploadutil.MultiFileProgress(fileIndex, fileTotal, fraction)
		fyne.Do(func() {
			progressBar.SetValue(overall)
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
