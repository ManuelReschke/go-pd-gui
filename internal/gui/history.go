package gui

import (
	"fmt"
	"image/color"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ManuelReschke/go-pd/pkg/pd"
)

// HistoryItems is a custom wrapper for the custom sort
type HistoryItems map[int]pd.FileGetUser

func (m HistoryItems) Len() int {
	return len(m)
}

func (m HistoryItems) Less(i, j int) bool {
	return m[j].DateUpload.Before(m[i].DateUpload)
}

func (m HistoryItems) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func BuildHistoryWindow() {
	window := MyApp.App.NewWindow("History")

	key := MyApp.App.Preferences().String(SettingAPIKey)

	req := &pd.RequestGetUserFiles{
		Auth: pd.Auth{
			APIKey: key,
		},
	}

	client := pd.New(nil, nil)
	rsp, err := client.GetUserFiles(req)
	if err != nil || rsp.StatusCode != 200 {
		text := container.New(layout.NewCenterLayout(), canvas.NewText("error occurred", color.White))
		window.SetContent(text)
		window.Resize(fyne.NewSize(600, 380))
		window.Show()
		return
	}

	// wrap all items to the new history map
	items := make(HistoryItems)
	for key, data := range rsp.Files {
		items[key] = data
	}

	// convert map to slice because map is not sortable
	keys := make([]int, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	// sort the slice after the custom rules you define earlier
	sort.Sort(items)

	list := widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(
				fmt.Sprintf(
					"%s | File: %s | Views: %d | DLs: %d",
					items[i].DateUpload.Format("2006-01-02 15:04:05"),
					items[i].Name,
					items[i].Views,
					items[i].Downloads,
				),
			)
		})

	line := canvas.NewLine(color.White)
	line.StrokeWidth = 5
	selectText := canvas.NewText("nothing selected", color.White)
	centerText := container.NewCenter(selectText)
	vbox := container.NewVBox(centerText, line)

	list.OnSelected = func(id widget.ListItemID) {
		// copy to clipboard
		fileURL := fmt.Sprintf("https://pixeldrain.com/u/%s", items[id].ID)
		MyApp.Window.Clipboard().SetContent(fileURL)

		// show animation for user
		selectText.Text = fmt.Sprintf("URL copied to the clipboard! File: %s", items[id].Name)

		go func() {
			selectText.Color = color.RGBA{R: 4, G: 139, B: 0, A: 1}
			selectText.Refresh()
			time.Sleep(3 * time.Second)
			selectText.Color = color.White
			selectText.Refresh()
		}()
	}

	window.SetContent(container.NewBorder(vbox, nil, nil, nil, container.NewMax(list)))
	window.Resize(fyne.NewSize(600, 380))
	window.Show()
}
