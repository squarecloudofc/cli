build:
	@go build -o bin/squarecloud cmd/squarecloud/main.go

.PHONY: completions
completions:
	mkdir -p completions
	bin/squarecloud completion bash >"completions/completions-bash.bash"
	bin/squarecloud completion fish >"completions/completions-fish.fish"
	bin/squarecloud completion zsh >"completions/completions-zsh.zsh"

.PHONY: install
install:
	@mkdir -p /usr/local/bin
	install bin/squarecloud /usr/local/bin

.PHONY: uninstall
uninstall:
	rm /usr/local/bin/squarecloud
