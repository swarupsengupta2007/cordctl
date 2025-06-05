APP_NAME := cordctl
MAIN_FILE := cordctl.go
VERSION := $(shell git describe --tags --always --dirty)
BUILD_DIR := dist

PLATFORMS := \
  linux/amd64 \
  linux/arm64 \
  linux/arm/v7 \
  linux/arm/v6 \
  darwin/amd64 \
  darwin/arm64 \
  windows/amd64 \
  windows/arm64

LDFLAGS := -s -w -X main.Version=$(VERSION)

.PHONY: build all clean compress

build:
	@echo "🔨 Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	@ext=$$( [ "$$(go env GOOS)" = "windows" ] && echo ".exe" || echo "" ); \
	outfile="$(BUILD_DIR)/$(APP_NAME)-$$(go env GOOS)-$$(go env GOARCH)$$ext"; \
	echo " -> $$outfile"; \
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $$outfile $(MAIN_FILE)

all:
	@echo "🌍 Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		$(MAKE) platform-build PLATFORM=$$platform || exit 1; \
	done

platform-build:
	@os=$(word 1,$(subst /, ,$(PLATFORM))); \
	 arch=$(word 2,$(subst /, ,$(PLATFORM))); \
	 variant=$(word 3,$(subst /, ,$(PLATFORM))); \
	 ext=$$( [ "$$os" = "windows" ] && echo ".exe" || echo "" ); \
	 dir="$$os-$$arch"; \
	 [ -n "$$variant" ] && dir="$$dir-$$variant"; \
	 outfile="$(BUILD_DIR)/$$dir/$(APP_NAME)-$$dir$$ext"; \
	 mkdir -p "$(BUILD_DIR)/$$dir"; \
	 echo " -> $$outfile"; \
	 buildcmd="GOOS=$$os GOARCH=$$arch CGO_ENABLED=0"; \
	 [ -n "$$variant" ] && buildcmd="$$buildcmd GOARM=$$variant"; \
	 buildcmd="$$buildcmd go build -ldflags '$(LDFLAGS)' -o $$outfile $(MAIN_FILE)"; \
	 echo $$buildcmd; \
	 sh -c "$$buildcmd" || echo "❌ Build failed for $$os/$$arch/$$variant"

clean:
	rm -rf $(BUILD_DIR)

compress:
	@which upx >/dev/null || (echo "⚠️  upx not installed"; exit 1)
	@find $(BUILD_DIR) -type f -executable -exec upx {} \;
