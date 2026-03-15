BINARY  := urai
CMD     := ./cmd/urai
SRC_DIR := $(CURDIR)/prose
BIN_DIR := $(CURDIR)/bin
VERSION := 0.1.0
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

SSH_BINARY := urai-ssh
SSH_CMD    := ./cmd/urai-ssh

.PHONY: build build-ssh run install release clean tidy

# Build for the current platform into the root directory
build:
	cd $(SRC_DIR) && go build $(LDFLAGS) -o $(CURDIR)/$(BINARY) $(CMD)

# Build the SSH server binary (cross-compile for Pi with: make build-ssh GOOS=linux GOARCH=arm64)
build-ssh:
	cd $(SRC_DIR) && go build $(LDFLAGS) -o $(CURDIR)/$(SSH_BINARY) $(SSH_CMD)

run: build
	$(CURDIR)/$(BINARY)

install:
	cd $(SRC_DIR) && go install $(LDFLAGS) $(CMD)

# Build release binaries for all platforms: bin/<os>/<arch>/urai[.exe]
release:
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		out=$(BIN_DIR)/$$os/$$arch/$(BINARY); \
		if [ "$$os" = "windows" ]; then out=$$out.exe; fi; \
		mkdir -p $(BIN_DIR)/$$os/$$arch; \
		echo "Building $$out ..."; \
		cd $(SRC_DIR) && GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$out $(CMD); \
	done

clean:
	rm -f $(CURDIR)/$(BINARY)
	rm -rf $(BIN_DIR)

tidy:
	cd $(SRC_DIR) && go mod tidy
