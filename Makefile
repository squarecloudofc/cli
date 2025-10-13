VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

BUILD_MAIN=cmd/squarecloud/main.go
BUILD_DIR=bin
BUILD_LDFLAGS=-ldflags="\
    -s -w \
    -X github.com/squarecloudofc/cli/internal/build.Version=$(VERSION) \
    -X github.com/squarecloudofc/cli/internal/build.Commit=$(COMMIT) \
    -X github.com/squarecloudofc/cli/internal/build.CommitDate=$(BUILD_TIME) \
"

INSTALL_DIR=$(HOME)/.squarecloud/bin

build:
	@go build $(BUILD_LDFLAGS) -o $(BUILD_DIR)/squarecloud $(BUILD_MAIN)

completions:
	mkdir -p $(BUILD_DIR)/completions
	@go run $(BUILD_DIR)/completions/squarecloud completion bash >"$(BUILD_DIR)/completions/completions-bash.bash"
	@go run $(BUILD_DIR)/completions/squarecloud completion fish >"$(BUILD_DIR)/completions/completions-fish.fish"
	@go run $(BUILD_DIR)/completions/squarecloud completion zsh >"$(BUILD_DIR)/completions/completions-zsh.zsh"

install: build
	@mkdir -p $(INSTALL_DIR)
	install bin/squarecloud $(INSTALL_DIR)

clean:
	rm -r bin/
