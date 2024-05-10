build:
	@go build -o bin/squarecloud cmd/squarecloud/main.go

install:
	@mkdir -p /usr/local/bin
	install bin/squarecloud /usr/local/bin

uninstall:
	rm /usr/local/bin/squarecloud
