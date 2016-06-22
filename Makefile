linux64:
	GOOS=linux GOARCH=amd64 go build -o dist/imageloader_linux

linux32:
	GOOS=linux GOARCH=386 go build -o dist/imageloader_linux

clean:
	rm -f dist/*
