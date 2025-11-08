package main

import (
	"encoding/json"
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Entry struct {
	URL      *widget.Entry `json:"url"`
	CONTRATO *widget.Entry `json:"contrato"`
	SERVER   *widget.Entry `json:"server"`
	DB       *widget.Entry `json:"db"`
	PASSWORD *widget.Entry `json:"password"`
	USER     *widget.Entry `json:"user"`
	PORT     *widget.Entry `json:"port"`
}

type fileJSON struct {
	URL      string `json:"url"`
	CONTRATO string `json:"contrato"`
	SERVER   string `json:"server"`
	DB       string `json:"db"`
	PASSWORD string `json:"password"`
	USER     string `json:"user"`
	PORT     string `json:"port"`
}

func main() {
	a := app.New()
	w := a.NewWindow("Perseo API")
	w.Resize(fyne.NewSize(600, 240))
	w.CenterOnScreen()
	w.SetFixedSize(true)

	credential := Entry{
		URL:      widget.NewEntry(),
		CONTRATO: widget.NewEntry(),
		SERVER:   widget.NewEntry(),
		DB:       widget.NewEntry(),
		PASSWORD: widget.NewEntry(),
		USER:     widget.NewEntry(),
		PORT:     widget.NewEntry(),
	}

	LoadCredentials(credential)

	inputJSON := widget.NewMultiLineEntry()
	inputJSON.SetPlaceHolder("BODY A ENCRIPTAR")

	inputBodySigned := widget.NewEntry()
	inputBodySigned.SetPlaceHolder("JSON ENCRIPTADO")
	inputBodySigned.Disable()

	generatedSignature := widget.NewEntry()
	generatedSignature.SetPlaceHolder("FIRMA GENERADA")
	generatedSignature.Disable()

	btnEncrypt := widget.NewButton("Encriptar 🔑", func() { Signed(inputJSON, inputBodySigned) })
	btnTest := widget.NewButton("Probar 🧪", func() { Fetch(w, inputBodySigned, generatedSignature) })
	btnClear := widget.NewButton("Limpiar 🗑️", func() { ClearField(inputJSON, inputBodySigned, generatedSignature) })
	btnConfig := widget.NewButton("Configurar ⚙️", func() { OpenModalConfiguration(w, credential) })

	buttons := container.NewVBox(btnEncrypt, btnTest, btnConfig, btnClear)

	btnCopyJSON := widget.NewButton("📋", func() { CopyResult(w, inputBodySigned.Text) })
	btnCopySignature := widget.NewButton("📋", func() { CopyResult(w, generatedSignature.Text) })

	jsonRow := container.NewBorder(nil, nil, nil, btnCopyJSON, inputBodySigned)
	signatureRow := container.NewBorder(nil, nil, nil, btnCopySignature, generatedSignature)

	top := container.NewBorder(nil, nil, nil, buttons, inputJSON)
	content := container.NewVBox(top, jsonRow, signatureRow)

	w.SetContent(content)
	w.ShowAndRun()
}

func OpenModalConfiguration(w fyne.Window, credential Entry) {

	credential.URL.SetPlaceHolder("http://localhost:8029/reportes_facturito_v2/proforma_individual")
	credential.CONTRATO.SetPlaceHolder("# DE CONTRATO")
	credential.SERVER.SetPlaceHolder("DOMINIO O IP")
	credential.DB.SetPlaceHolder("BASE DE DATOS")
	credential.PASSWORD.SetPlaceHolder("CONTRASEÑA DB")
	credential.USER.SetPlaceHolder("USUARIO DB")
	credential.PORT.SetPlaceHolder("PUERTO DB")

	var popup *widget.PopUp

	btnCerrar := widget.NewButton("CERRAR", func() { popup.Hide() })
	btnGuardar := widget.NewButton("GUARDAR", func() { SaveCredentials(popup, credential) })

	buttonRow := container.NewHBox(btnCerrar, btnGuardar)

	grid := container.NewGridWithColumns(3,
		container.NewVBox(credential.CONTRATO, credential.DB),
		container.NewVBox(credential.SERVER, credential.PASSWORD),
		container.NewVBox(credential.USER, credential.PORT),
	)

	content := container.NewVBox(
		container.NewCenter(
			widget.NewLabelWithStyle("CONFIGURACIÓN", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		),
		credential.URL,
		grid,
		layout.NewSpacer(),
		container.NewCenter(buttonRow),
	)

	popup = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	popup.Resize(fyne.NewSize(650, 250))
	popup.Show()
}

func OpenModalFetching(w fyne.Window, json *widget.Entry) {

	responseJSON := widget.NewMultiLineEntry()
	responseJSON.SetText(json.Text)
	responseJSON.SetMinRowsVisible(7) // 🔹 más alto

	var popup *widget.PopUp

	btnCerrar := widget.NewButton("CERRAR", func() {
		popup.Hide()
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("RESPUESTA DEL SERVIDOR", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		responseJSON,
		container.NewCenter(btnCerrar),
	)

	popup = widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
	popup.Resize(fyne.NewSize(900, 400))
	popup.Show()

}

func ClearField(inputs ...*widget.Entry) {
	for _, entry := range inputs {
		entry.SetText("")
	}
}

func Signed(inputBodySigned, generatedSignature *widget.Entry) {
	generatedSignature.SetText(inputBodySigned.Text)
}

func Fetch(w fyne.Window, inputBodySigned, generatedSignature *widget.Entry) {
	OpenModalFetching(w, inputBodySigned)
}

func SaveCredentials(p *widget.PopUp, credential Entry) {
	data := fileJSON{
		URL:      credential.URL.Text,
		CONTRATO: credential.CONTRATO.Text,
		SERVER:   credential.SERVER.Text,
		DB:       credential.DB.Text,
		PASSWORD: credential.PASSWORD.Text,
		USER:     credential.USER.Text,
		PORT:     credential.PORT.Text,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error al convertir a JSON:", err)
		return
	}

	err = os.WriteFile("config.json", jsonBytes, 0644)
	if err != nil {
		fmt.Println("Error al guardar el archivo:", err)
		return
	}

	fmt.Println("Configuración guardada correctamente en config.json")

	p.Hide()

	LoadCredentials(credential)
}

func LoadCredentials(credential Entry) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("No se pudo leer config.json:", err)
		return
	}

	var c fileJSON
	err = json.Unmarshal(data, &c)
	if err != nil {
		fmt.Println("Error al parsear config.json:", err)
		return
	}

	credential.URL.SetText(c.URL)
	credential.CONTRATO.SetText(c.CONTRATO)
	credential.SERVER.SetText(c.SERVER)
	credential.DB.SetText(c.DB)
	credential.PASSWORD.SetText(c.PASSWORD)
	credential.USER.SetText(c.USER)
	credential.PORT.SetText(c.PORT)

	fmt.Println("Configuración cargada correctamente ✅")
}

func CopyResult(w fyne.Window, text string) {
	if text != "" {
		w.Clipboard().SetContent(text)
		dialog.ShowInformation("Copiado ✅", "El texto se ha copiado al portapapeles.", w)
	} else {
		dialog.ShowInformation("Vacío ⚠️", "No hay texto para copiar.", w)
	}
}
