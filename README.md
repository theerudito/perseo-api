go build -ldflags="-H=windowsgui" -o perseo-api.exe
fyne package -os windows -icon icon.png
