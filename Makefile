.PHONY: all build install uninstall clean

BINARY_NAME=zet-terminal
BIN_DIR=bin
LOCAL_BIN=$(HOME)/.local/bin
DESKTOP_DIR=$(HOME)/.local/share/applications
KDE5_SERVICE_DIR=$(HOME)/.local/share/kservices5/ServiceMenus
KDE6_SERVICE_DIR=$(HOME)/.local/share/kio/servicemenus

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/zet-terminal

install: build
	@echo "Installing $(BINARY_NAME) to $(LOCAL_BIN)..."
	@mkdir -p $(LOCAL_BIN)
	cp -f $(BIN_DIR)/$(BINARY_NAME) $(LOCAL_BIN)/$(BINARY_NAME)
	@chmod +x $(LOCAL_BIN)/$(BINARY_NAME)

	@echo "Installing application launcher..."
	@mkdir -p $(DESKTOP_DIR)
	cp -f $(BINARY_NAME).desktop $(DESKTOP_DIR)/$(BINARY_NAME).desktop
	sed -i 's|Exec=zet-terminal|Exec=$(LOCAL_BIN)/zet-terminal|g' $(DESKTOP_DIR)/$(BINARY_NAME).desktop
	@chmod +x $(DESKTOP_DIR)/$(BINARY_NAME).desktop

	@echo "Installing Dolphin context menu actions..."
	@mkdir -p $(KDE5_SERVICE_DIR)
	cp -f $(BINARY_NAME)-action.desktop $(KDE5_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	sed -i 's|Exec=zet-terminal|Exec=$(LOCAL_BIN)/zet-terminal|g' $(KDE5_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	@chmod +x $(KDE5_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	@mkdir -p $(KDE6_SERVICE_DIR)
	cp -f $(BINARY_NAME)-action.desktop $(KDE6_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	sed -i 's|Exec=zet-terminal|Exec=$(LOCAL_BIN)/zet-terminal|g' $(KDE6_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	@chmod +x $(KDE6_SERVICE_DIR)/$(BINARY_NAME)-action.desktop

	@echo "Updating desktop database..."
	@if command -v update-desktop-database >/dev/null 2>&1; then \
		update-desktop-database $(DESKTOP_DIR) >/dev/null 2>&1 || true; \
	fi

	@echo "Rebuilding KDE sycoca database..."
	@if command -v kbuildsycoca6 >/dev/null 2>&1; then \
		kbuildsycoca6 >/dev/null 2>&1 || true; \
	elif command -v kbuildsycoca5 >/dev/null 2>&1; then \
		kbuildsycoca5 >/dev/null 2>&1 || true; \
	fi

	@echo "Installation complete!"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(LOCAL_BIN)/$(BINARY_NAME)
	rm -f $(DESKTOP_DIR)/$(BINARY_NAME).desktop
	rm -f $(KDE5_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	rm -f $(KDE6_SERVICE_DIR)/$(BINARY_NAME)-action.desktop
	@if command -v update-desktop-database >/dev/null 2>&1; then \
		update-desktop-database $(DESKTOP_DIR) >/dev/null 2>&1 || true; \
	fi
	@echo "Uninstallation complete!"

clean:
	@echo "Cleaning up build artifacts..."
	rm -rf $(BIN_DIR)
