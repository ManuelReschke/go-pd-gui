package gui

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ManuelReschke/go-pd/pkg/pd"
	"image/color"
	"strings"
)

const (
	SettingAPIKey           = "setting.apikey"
	SettingUserInfoJSON     = "setting.user.info.json"
	SettingUsername         = "setting.user.name"
	SettingUserSubscription = "setting.user.subscription"
)

func BuildSettingsWindow() {
	w := MyApp.App.NewWindow("Settings")

	// FORM INPUT API-KEY
	formLabelAPIKey := canvas.NewText(FormLabel, color.White)

	inputAPIKey := widget.NewEntry()
	inputAPIKey.SetPlaceHolder(FormLabelInput)
	//MyApp.Settings.APIKey = binding.NewString()
	_ = MyApp.Settings.APIKey.Set(MyApp.App.Preferences().StringWithFallback(SettingAPIKey, ""))
	inputAPIKey.Bind(MyApp.Settings.APIKey)

	userData := getUserDataFromSettings()
	displayUsername := userData.Username
	if displayUsername != "" {
		_ = MyApp.Settings.Username.Set(displayUsername)
	}
	if displayUsername == "" {
		displayUsername = "Invalid API Key or PixelDrain is not available"
	}
	userSettingsContainer := canvas.NewText("User: "+displayUsername, color.White)
	displaySubscription := userData.Subscription.Name
	if displaySubscription != "" {
		_ = MyApp.Settings.Subscription.Set(displaySubscription)
	}
	displaySubscription, _ = MyApp.Settings.Subscription.Get()
	userSettingsContainer2 := canvas.NewText("Subscription: "+displaySubscription, color.White)
	vertBox := container.NewVBox(userSettingsContainer, layout.NewSpacer(), userSettingsContainer2)

	containerForm := container.New(layout.NewFormLayout(), formLabelAPIKey, inputAPIKey, layout.NewSpacer(), vertBox)
	containerForm.Resize(fyne.NewSize(0, 200))

	saveButton := widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), func() {
		key, _ := MyApp.Settings.APIKey.Get()

		// check if the key is valid fetch data from pixeldrain /user
		userInfos := ""
		req := &pd.RequestGetUser{
			Auth: pd.Auth{
				APIKey: key,
			},
		}

		client := pd.New(nil, nil)
		rsp, err := client.GetUser(req)
		if err != nil || rsp.StatusCode != 200 {
			// nothing
		} else {
			b, _ := json.Marshal(rsp)
			userInfos = string(b)
		}

		MyApp.App.Preferences().SetString(SettingUserInfoJSON, userInfos)
		userData := getUserDataFromSettings()
		MyApp.App.Preferences().SetString(SettingUsername, userData.Username)
		MyApp.App.Preferences().SetString(SettingUserSubscription, userData.Subscription.Name)
		_ = MyApp.Settings.Username.Set(userData.Username)
		_ = MyApp.Settings.Subscription.Set(userData.Subscription.Name)

		MyApp.App.Preferences().SetString(SettingAPIKey, strings.TrimSpace(key)) // store user input
		w.Hide()
	})
	containerButton := container.New(layout.NewCenterLayout(), saveButton)

	content := container.New(layout.NewVBoxLayout(), containerForm, layout.NewSpacer(), containerButton)

	w.SetContent(content)
	w.Resize(fyne.NewSize(480, 380))
	w.Show()
}

func getUserDataFromSettings() *pd.ResponseGetUser {
	userSettingJson := MyApp.App.Preferences().StringWithFallback(SettingUserInfoJSON, "")

	userData := &pd.ResponseGetUser{}
	err := json.Unmarshal([]byte(userSettingJson), userData)
	if err != nil {
		return userData
	}

	return userData
}
