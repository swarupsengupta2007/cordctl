APP_NAME := cordctl
VERSION  := $(shell git describe --tags --always --dirty)
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

.PHONY: all build clean compress

all: build

build:
	@echo "Building $(APP_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach platform,$(PLATFORMS),\
		$(call build_platform,$(platform)))

define build_platform
  os=$(word 1,$(subst /, ,$1)); \
  arch=$(word 2,$(subst /, ,$1)); \
  variant=$(word 3,$(subst /, ,$1)); \
  ext=$$( [ "$$os" = "windows" ] && echo ".exe" || echo "" ); \
  outdir="$(BUILD_DIR)/$$os-$$arch"; \
  outfile="$(APP_NAME)-$$os-$$arch"; \
  [ -n "$$variant" ] && outdir="$$outdir-$$variant" && outfile="$$outfile-$$variant"; \
  mkdir -p $$outdir; \
  echo " -> $$outdir/$$outfile$$ext"; \
  env_vars="GOOS=$$os GOARCH=$$arch CGO_ENABLED=0"; \
  [ -n "$$variant" ] && env_vars="$$env_vars GOARM=$$variant"; \
  sh -c "$$env_vars go build -ldflags '$(LDFLAGS)' -o $$outdir/$$outfile$$ext cordctl.go"
endef

clean:
	rm -rf $(BUILD_DIR)

compress:
	@which upx >/dev/null || (echo "upx not installed"; exit 1)
	@find $(BUILD_DIR) -type f -executable -exec upx {} \;
