build:
	@go build -o bin/squarecloud cmd/squarecloud/main.go

.PHONY: completions
completions:
	mkdir -p completions
	@go run ./cmd/squarecloud completion bash >"completions/completions-bash.bash"
	@go run ./cmd/squarecloud completion fish >"completions/completions-fish.fish"
	@go run ./cmd/squarecloud completion zsh >"completions/completions-zsh.zsh"

.PHONY: install
install:
	@mkdir -p /usr/local/bin
	install bin/squarecloud /usr/local/bin

.PHONY: uninstall
uninstall:
	rm /usr/local/bin/squarecloud

.PHONY: clean
clean:
	rm -r bin/ completions/ dist/
