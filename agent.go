package main

// NewApp creates a new App application struct
func NewAgent() *App {
	return &App{}
}

func (a *App) OpenNewWindow() {
	// Create a new window
	// 	app := wails.CreateApp(appOptions)
	//   mainWindow := app.CreateWindow(windowOptions)
	//   childWindow := mainWindow.CreateWindow(windowOptions)
	// newWindow, err := wails.newWindow(&wails.CreateWindowOptions{
	// 	Width:  800,
	// 	Height: 600,
	// 	Title:  "New Window",
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// // Load your front-end content for the new window
	// newWindow.SetContent("<html><body><h1>New Window</h1></body></html>")

	// // Show the window
	// newWindow.Show()
}
