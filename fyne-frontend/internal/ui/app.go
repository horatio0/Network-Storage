package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"reverseproxy-poc/internal/client"
)

// RunApp initializes and runs the main Fyne application.
func RunApp() {
	myApp := app.NewWithID("com.networkstorage.client")
	myApp.Settings().SetTheme(NewCustomTheme())

	myWindow := myApp.NewWindow("Network Storage Control")
	httpClient := client.NewHTTPClient(myApp)
	SetupMainWindow(myApp, myWindow, httpClient)

	myWindow.Resize(fyne.NewSize(1000, 600))
	myWindow.ShowAndRun()
}
