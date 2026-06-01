#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# SpecAI Installation Script
REPO="KevG1t/SpecAI"
BINARY_NAME="specai"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Architecture
detect_platform() {
    local os
    local arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*) os="windows";;
        *)          echo -e "${RED}Unsupported OS: $(uname -s)${NC}" && exit 1;;
    esac
    
    # Detect Architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64";;
        arm64|aarch64) arch="arm64";;
        armv7l) arch="arm";;
        i386) arch="386";;
        *) echo -e "${RED}Unsupported architecture: $(uname -m)${NC}" && exit 1;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install binary
install_specai() {
    local platform=$(detect_platform)
    local version=$(get_latest_version)
    
    if [ -z "$version" ]; then
        echo -e "${RED}Failed to get latest version${NC}"
        exit 1
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
    
    echo -e "${YELLOW}Installing SpecAI ${version} for ${platform}...${NC}"
    
    # Create temporary file
    local temp_file=$(mktemp)
    
    # Download binary
    echo "Downloading from: ${download_url}"
    if ! curl -fsSL "${download_url}" -o "${temp_file}"; then
        echo -e "${RED}Failed to download SpecAI binary${NC}"
        rm -f "${temp_file}"
        exit 1
    fi
    
    # Make executable
    chmod +x "${temp_file}"
    
    # Install binary
    if [ -w "${INSTALL_DIR}" ]; then
        mv "${temp_file}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo "Installing to ${INSTALL_DIR} (requires sudo)..."
        sudo mv "${temp_file}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    echo -e "${GREEN}✅ SpecAI installed successfully!${NC}"
    echo -e "${GREEN}Run 'specai' to get started${NC}"
}

# Check if binary already exists
check_existing() {
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        local current_version=$("${BINARY_NAME}" --version 2>/dev/null | head -n1 || echo "unknown")
        echo -e "${YELLOW}SpecAI is already installed: ${current_version}${NC}"
        read -p "Do you want to reinstall? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation cancelled."
            exit 0
        fi
    fi
}

# Main installation function
main() {
    echo -e "${GREEN}🚀 SpecAI Installation Script${NC}"
    echo "================================"
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        echo -e "${RED}curl is required but not installed${NC}"
        exit 1
    fi
    
    check_existing
    install_specai
    
    # Verify installation
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        echo -e "${GREEN}Installation verified successfully!${NC}"
        echo ""
        echo "Quick start:"
        echo "  specai              # Interactive mode"
        echo "  specai setup cursor # Setup Cursor rules"
        echo "  specai --help       # Show all options"
    else
        echo -e "${RED}Installation failed - binary not found in PATH${NC}"
        exit 1
    fi
}

# Run main function
main "$@"