# ==== USER PATHS ====
WHISPER_ROOT := /Users/illashisko/Documents/Code/whisper.cpp
BUILD_DIR    := $(WHISPER_ROOT)/build_go

# ==== DERIVED PATHS ====
INCLUDE_PATH := $(WHISPER_ROOT)/include:$(WHISPER_ROOT)/ggml/include
LIBRARY_PATH := $(BUILD_DIR)/src:$(BUILD_DIR)/ggml/src:$(BUILD_DIR)/ggml/src/ggml-blas:$(BUILD_DIR)/ggml/src/ggml-metal

# для metal нужно указать путь к мет. ресурсам (папка проекта whisper.cpp)
GGML_METAL_PATH_RESOURCES := $(WHISPER_ROOT)

# macOS фреймворки и доп. либы
EXT_LDFLAGS := -framework Foundation -framework Metal -framework MetalKit -lggml-metal -lggml-blas

# ==== GO APP ====
APP_NAME := app
PKG      := ./cmd

# ==== PHONY ====
.PHONY: all whisper build run clean tidy

# ==== DEFAULT ====
all: whisper build

# ==== BUILD WHISPER LIB ====
whisper:
	@echo "==> Configure & build whisper.cpp (static lib)"
	@cmake -S $(WHISPER_ROOT) -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF
	@cmake --build $(BUILD_DIR) --target whisper

# ==== BUILD GO BINARY ====
build: tidy
	@echo "==> Build Go app"
	@C_INCLUDE_PATH=$(INCLUDE_PATH) \
	LIBRARY_PATH=$(LIBRARY_PATH) \
	GGML_METAL_PATH_RESOURCES=$(GGML_METAL_PATH_RESOURCES) \
	go build -ldflags "-extldflags '$(EXT_LDFLAGS)'" -o $(APP_NAME) $(PKG)

# ==== RUN ====
run: build
	@./$(APP_NAME)

# ==== GO MOD TIDY ====
tidy:
	@go mod tidy

# ==== CLEAN ====
clean:
	@echo "==> Clean Go build"
	@rm -f $(APP_NAME)
	@go clean
