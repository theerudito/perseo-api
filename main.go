package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Perseo API")

	w.Resize(fyne.NewSize(600, 400))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	w.SetContent(widget.NewLabel("Hello World!"))
	w.ShowAndRun()
}
