all: win64 win32 linux64 linux32 darwin

install: 
	go install

win64:
	GOOS=windows GOARCH=amd64 go build -o dist/imageloader.exe

win32:
	GOOS=windows GOARCH=386 go build -o dist/imageloader.exe

linux64:
	GOOS=linux GOARCH=amd64 go build -o dist/imageloader_linux_amd64

linux32:
	GOOS=linux GOARCH=386 go build -o dist/imageloader_linux_i386

darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/imageloader_darwin

clean:
	rm -f dist/*
