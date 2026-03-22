APP_NAME := wol-server
BUILD_DIR := build
CMD_PATH := ./cmd/app
OUTPUT := $(BUILD_DIR)/$(APP_NAME)
TARGET_OS ?= linux
TARGET_ARCH ?= arm64
CGO_ENABLED ?= 1
TARGET_CC ?= aarch64-linux-gnu-gcc
BUILD_FLAGS ?= -trimpath -ldflags="-s -w"
UPX_ENABLED ?= 1
UPX_FLAGS ?= --best --lzma
OPENWRT_CC := aarch64-openwrt-linux-musl-gcc
OPENWRT_SDK ?=
OPENWRT_STAGING_DIR ?= $(if $(OPENWRT_SDK),$(OPENWRT_SDK)/staging_dir,)
OPENWRT_TOOLCHAIN_BIN ?= $(firstword $(wildcard $(OPENWRT_STAGING_DIR)/toolchain-aarch64_cortex-a53_gcc-13.3.0_musl/bin))
OPENWRT_CC_WRAPPER := $(BUILD_DIR)/openwrt-cc

.PHONY: check-deps check-openwrt-sdk build build-openwrt

check-deps:
	@if ! command -v $(TARGET_CC) >/dev/null 2>&1; then \
		if [ "$(TARGET_CC)" = "aarch64-linux-gnu-gcc" ]; then \
			echo "Installing gcc-aarch64-linux-gnu..."; \
			if command -v apt-get >/dev/null 2>&1; then \
				if command -v sudo >/dev/null 2>&1; then \
					sudo apt-get update && sudo apt-get install -y gcc-aarch64-linux-gnu; \
				else \
					apt-get update && apt-get install -y gcc-aarch64-linux-gnu; \
				fi; \
			else \
				echo "Unsupported package manager. Install $(TARGET_CC) manually."; \
				exit 1; \
			fi; \
		else \
			echo "Compiler $(TARGET_CC) not found."; \
			echo "Install OpenWrt SDK/toolchain manually and export TARGET_CC=$(TARGET_CC)."; \
			exit 1; \
		fi; \
	fi
	@if [ "$(UPX_ENABLED)" = "1" ]; then \
		if ! command -v upx >/dev/null 2>&1; then \
			echo "Installing upx..."; \
			if command -v apt-get >/dev/null 2>&1; then \
				if command -v sudo >/dev/null 2>&1; then \
					sudo apt-get update && (sudo apt-get install -y upx-ucl || sudo apt-get install -y upx); \
				else \
					apt-get update && (apt-get install -y upx-ucl || apt-get install -y upx); \
				fi; \
			else \
				echo "Unsupported package manager. Install upx manually."; \
				exit 1; \
			fi; \
		fi; \
	fi

build: check-deps
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CC=$(TARGET_CC) go build $(BUILD_FLAGS) -o $(OUTPUT) $(CMD_PATH)
	@if [ "$(UPX_ENABLED)" = "1" ]; then upx $(UPX_FLAGS) $(OUTPUT); fi

check-openwrt-sdk:
	@if [ -z "$(OPENWRT_SDK)" ]; then \
		echo "OPENWRT_SDK is not set."; \
		echo "Usage: make build-openwrt OPENWRT_SDK=/path/to/openwrt-sdk-* [UPX_ENABLED=0]"; \
		exit 1; \
	fi
	@if [ ! -d "$(OPENWRT_SDK)" ]; then \
		echo "OPENWRT_SDK path does not exist: $(OPENWRT_SDK)"; \
		exit 1; \
	fi
	@if [ -z "$(OPENWRT_TOOLCHAIN_BIN)" ] || [ ! -d "$(OPENWRT_TOOLCHAIN_BIN)" ]; then \
		echo "OpenWrt toolchain bin dir not found under $(OPENWRT_STAGING_DIR)/toolchain-*/bin"; \
		exit 1; \
	fi

build-openwrt: check-openwrt-sdk
	@PATH="$(OPENWRT_TOOLCHAIN_BIN):$$PATH" STAGING_DIR="$(OPENWRT_STAGING_DIR)" \
		$(MAKE) check-deps TARGET_CC=$(OPENWRT_CC) UPX_ENABLED=$(UPX_ENABLED)
	mkdir -p $(BUILD_DIR)
	printf '%s\n' '#!/usr/bin/env sh' 'export STAGING_DIR="$(OPENWRT_STAGING_DIR)"' \
		'exec "$(OPENWRT_TOOLCHAIN_BIN)/$(OPENWRT_CC)" "$$@"' > $(OPENWRT_CC_WRAPPER)
	chmod +x $(OPENWRT_CC_WRAPPER)
	PATH="$(OPENWRT_TOOLCHAIN_BIN):$$PATH" \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CC="$(abspath $(OPENWRT_CC_WRAPPER))" \
		go build $(BUILD_FLAGS) -o $(OUTPUT) $(CMD_PATH)
	@if [ "$(UPX_ENABLED)" = "1" ]; then \
		STAGING_DIR="$(OPENWRT_STAGING_DIR)" PATH="$(OPENWRT_TOOLCHAIN_BIN):$$PATH" \
		upx $(UPX_FLAGS) $(OUTPUT); \
	fi
