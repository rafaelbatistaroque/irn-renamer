# Nome base do aplicativo
APP_NAME := irn-renamer

# Diretório de saída para os binários
OUTPUT_DIR := bin

# Flags do linker para reduzir o tamanho do binário
LDFLAGS := -ldflags="-s -w"

# Define os nomes dos arquivos de saída alvo
TARGET_WIN_AMD64 := $(OUTPUT_DIR)/$(APP_NAME).exe
TARGET_MAC_ARM64 := $(OUTPUT_DIR)/$(APP_NAME)

# --- Targets Principais ---

# Target padrão: executa ao rodar apenas 'make'. Constrói ambos os alvos.
.PHONY: all
all: $(TARGET_WIN_AMD64) $(TARGET_MAC_ARM64)

# Target para construir para Windows (amd64)
# O nome do target é o próprio arquivo que queremos gerar.
$(TARGET_WIN_AMD64):
	@echo "--> Building for Windows AMD64..."
	@mkdir -p $(OUTPUT_DIR) # Garante que o diretório de saída exista
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(TARGET_WIN_AMD64) .

# Target para construir para macOS (Apple Silicon arm64)
# O nome do target é o próprio arquivo que queremos gerar.
$(TARGET_MAC_ARM64):
	@echo "--> Building for macOS ARM64 (Apple Silicon)..."
	@mkdir -p $(OUTPUT_DIR) # Garante que o diretório de saída exista
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(TARGET_MAC_ARM64) .

# Target para limpar os artefatos de build (o diretório bin/)
.PHONY: clean
clean:
	@echo "--> Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)

# --- Targets de Conveniência (Opcional) ---
# Permitem chamar 'make windows' ou 'make mac-arm' explicitamente

.PHONY: windows mac-arm
windows: $(TARGET_WIN_AMD64)
mac-arm: $(TARGET_MAC_ARM64)