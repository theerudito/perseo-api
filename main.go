package main

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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
	inputJSON.SetPlaceHolder("JSON A ENCRIPTAR")

	inputBodySigned := widget.NewEntry()
	inputBodySigned.SetPlaceHolder("JSON ENCRIPTADO")
	inputBodySigned.Disable()

	inputResponse := widget.NewMultiLineEntry()
	inputResponse.SetPlaceHolder("RESPUESTA DEL SERVIDOR")
	inputResponse.SetMinRowsVisible(7)

	inputSignature := widget.NewEntry()
	inputSignature.SetPlaceHolder("FIRMA GENERADA")
	inputSignature.Disable()

	btnEncrypt := widget.NewButton("Encriptar 🔑", func() {
		Signed(credential.CONTRATO, credential.DB, credential.URL, inputJSON, inputBodySigned, inputSignature, w)
	})

	btnTest := widget.NewButton("Probar 🧪", func() {
		Fetch(w, inputSignature, inputBodySigned, inputResponse, credential.URL, credential.USER, credential.SERVER, credential.PASSWORD, credential.PORT, credential.DB)
	})

	btnClear := widget.NewButton("Limpiar 🗑️", func() { ClearField(inputJSON, inputBodySigned, inputSignature, inputResponse) })
	btnConfig := widget.NewButton("Configurar ⚙️", func() { OpenModalConfiguration(w, credential) })

	buttons := container.NewVBox(btnEncrypt, btnTest, btnConfig, btnClear)

	btnCopyJSON := widget.NewButton("📋", func() { CopyResult(w, inputBodySigned.Text) })
	btnCopySignature := widget.NewButton("📋", func() { CopyResult(w, inputSignature.Text) })

	jsonRow := container.NewBorder(nil, nil, nil, btnCopyJSON, inputBodySigned)
	signatureRow := container.NewBorder(nil, nil, nil, btnCopySignature, inputSignature)

	top := container.NewBorder(nil, nil, nil, buttons, inputJSON)
	content := container.NewVBox(top, jsonRow, signatureRow, inputResponse)

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

func ClearField(inputs ...*widget.Entry) {
	for _, entry := range inputs {
		entry.SetText("")
	}
}

func Signed(inputContrato, inputDB, inputURL, inputJSON, inputBodySigned, inputSignature *widget.Entry, w fyne.Window) {

	clave := GenerarClave(inputContrato.Text)

	claveTemporal, _ := EncriptarSHA256(inputDB.Text+"-"+"RICARDO", clave)

	serializarJson, err := SerializarJSON(inputJSON.Text)
	if err != nil {
		dialog.ShowInformation("❌ Error", "JSON INVÁLIDO", w)
		return
	}

	jsonCifrado, _ := EncriptarSHA256(serializarJson, clave)

	bodyHash := HashearBody(jsonCifrado)

	timeStamp := GenerarTimestamp()

	urlNorm := NormalizarURL(inputURL.Text)

	firma := GenerarFirma(urlNorm, claveTemporal, timeStamp, bodyHash, clave)

	autorizacion := GenerarAutorizacion(timeStamp, firma, claveTemporal+"*"+clave)

	inputBodySigned.SetText(jsonCifrado)
	inputSignature.SetText(autorizacion)

}

func Fetch(w fyne.Window, inputSignature, inputBodySigned, textRespuesta, inputURI, inputUsuario, inputServidor, inputPassword, inputPuerto, inputDB *widget.Entry) {

	if inputSignature.Text == "" || inputBodySigned.Text == "" {
		dialog.ShowInformation("❌ Error", "Faltan datos obligatorios", w)
		return
	}

	server := strings.TrimSpace(strings.Trim(inputServidor.Text, "/"))
	url := NormalizarURL(inputURI.Text)

	headers := HeaderForApi(
		inputSignature.Text,
		url,
		inputUsuario.Text,
		server,
		inputPassword.Text,
		inputPuerto.Text,
		inputDB.Text,
	)

	for k, v := range headers {
		fmt.Printf("   %s: %s\n", k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}

	body := bytes.NewBuffer([]byte(inputBodySigned.Text))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		dialog.ShowInformation("❌ Error", fmt.Sprintf("Error creando request: %v", err), w)
		return
	}

	for k, v := range headers {
		if strings.EqualFold(k, "Content-Type") {
			continue
		}
		req.Header.Add(k, v)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		dialog.ShowInformation("❌ Error", fmt.Sprintf("Error en la petición: %v", err), w)
		return
	}

	defer resp.Body.Close()

	fmt.Println("👉 StatusCode:", resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	respuesta := string(bodyBytes)

	fmt.Println("👉 Respuesta del servidor:")
	fmt.Println(respuesta)

	textRespuesta.Disable()
	textRespuesta.SetText(fmt.Sprintf("HTTP %d\n\n%s", resp.StatusCode, respuesta))
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

func GenerarClave(clave string) string {
	hash := sha256.Sum256([]byte(clave))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize

	padtext := make([]byte, padding)

	for i := range padtext {
		padtext[i] = byte(padding)
	}

	return append(data, padtext...)
}

func EncriptarSHA256(dato string, clave string) (string, error) {

	datoBytes := []byte(dato)
	claveBytes := []byte(clave)

	key := sha256.Sum256(claveBytes)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	datoBytes = PKCS7Padding(datoBytes, blockSize)

	encrypted := make([]byte, len(datoBytes))
	for bs, be := 0, blockSize; bs < len(datoBytes); bs, be = bs+blockSize, be+blockSize {
		block.Encrypt(encrypted[bs:be], datoBytes[bs:be])
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func SerializarJSON(jsonStr string) (string, error) {

	var cleanedLines []string
	lines := strings.Split(jsonStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") {
			continue
		}
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleanJSON := strings.Join(cleanedLines, "\n")

	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(cleanJSON)); err != nil {
		return "", err
	}
	return buf.String(), nil

}

func HashearBody(json string) string {

	normalizado := strings.TrimSpace(strings.ReplaceAll(json, "\r\n", "\n"))
	dataBytes := []byte(normalizado)
	hashBytes := sha256.Sum256(dataBytes)
	return base64.StdEncoding.EncodeToString(hashBytes[:])

}

func GenerarTimestamp() string {
	return time.Now().Format("20060102T150405")
}

func GenerarFirma(uri string, claveTemporal, timestamp, bodyHash, clave string) string {

	stringToSign := "POST" + "," + uri + "," + claveTemporal + "," + timestamp + "," + bodyHash

	stringToSign = strings.NewReplacer(
		"/", "",
		",", "",
		"_", "",
		"<", "",
		">", "",
		"=", "",
	).Replace(stringToSign)

	// Usa la clave principal (igual que C#)
	h := hmac.New(sha256.New, []byte(clave))
	h.Write([]byte(stringToSign))
	firma := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return firma
}

func GenerarAutorizacion(timeStamp string, firma string, clave string) string {
	return fmt.Sprintf(
		"AWS4-HMAC-SHA256 AccessKeyId=%s, Timestamp=%s, Signature=%s",
		clave, timeStamp, firma)
}

func HeaderForApi(authorization, uri, user, server, password, puerto, dbName string) map[string]string {
	headers := map[string]string{
		"Accept":             "/",
		"AutorizacionPerseo": authorization,
		"Method":             "POST",
		"URI":                uri,
		"SUSUARIO":           user,
		"SPASS":              password,
		"SPORT":              puerto,
		"SSERVER":            server,
		"DB":                 dbName,
		"Content-Type":       "application/json",
	}
	return headers
}

func NormalizarURL(raw string) string {
	return strings.TrimSpace(strings.Trim(raw, "/"))
}
