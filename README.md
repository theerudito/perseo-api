go build -ldflags="-H=windowsgui" -o perseo-api.exe
fyne package -os windows -icon icon.png

OR 

go >= 1.13
docker

go install github.com/fyne-io/fyne-cross@latest

fyne-cross windows -arch=*
fyne-cross windows -arch=amd64,386
