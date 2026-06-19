package ui

import (
	"reverseproxy-poc/internal/client"
	"reverseproxy-poc/internal/webrtc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// RunApp initializes and runs the main Fyne application.
func RunApp() {
	myApp := app.NewWithID("com.networkstorage.client")
	myApp.Settings().SetTheme(NewCustomTheme())

	myWindow := myApp.NewWindow("Network Storage Control")
	httpClient := client.NewHTTPClient(myApp)
	SetupMainWindow(myApp, myWindow, httpClient)

	go startListenerLoop(myApp)

	myWindow.Resize(fyne.NewSize(1000, 600))
	myWindow.ShowAndRun()
}

func startListenerLoop(a fyne.App) {
	ip := a.Preferences().StringWithFallback("server_ip", "")
	port := a.Preferences().StringWithFallback("server_port", "8080")
	if ip != "" {
		webrtc.StartBackgroundListener(ip, port)
	}
}
