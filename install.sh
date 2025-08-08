#!/bin/bash

# Ask Terminal Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/jayaprakash-shanmugam/Ask-Terminal/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="askterminal"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/askterminal"
REPO_URL="https://github.com/jayaprakash-shanmugam/Ask-Terminal"

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.19 or higher."
        echo "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_status "Found Go version: $GO_VERSION"
}

# Install the binary
install_binary() {
    print_status "Installing Ask Terminal..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Clone and build
    print_status "Downloading source code..."
    git clone "$REPO_URL.git" || {
        print_error "Failed to clone repository"
        exit 1
    }
    
    cd Ask-Terminal
    print_status "Building binary..."
    go mod tidy
    go build -o "$BINARY_NAME" main.go
    
    # Install to system path
    print_status "Installing to $INSTALL_DIR..."
    if sudo mv "$BINARY_NAME" "$INSTALL_DIR/"; then
        print_success "Binary installed to $INSTALL_DIR/$BINARY_NAME"
    else
        print_error "Failed to install binary. Try running with sudo."
        exit 1
    fi
    
    # Clean up
    cd /
    rm -rf "$TEMP_DIR"
}

# Setup API key
setup_api_key() {
    if [[ -n "$GEMINI_API_KEY" ]]; then
        print_success "Gemini API key already configured"
        return 0
    fi
    
    print_warning "Gemini API key not found in environment"
    echo
    echo "To use AI features, you'll need a free Gemini API key:"
    echo "1. Visit: https://makersuite.google.com/app/apikey"
    echo "2. Create a new API key (free)"
    echo "3. Copy the key"
    echo
    
    read -p "Do you have a Gemini API key? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter your Gemini API key: " -rs API_KEY
        echo
        
        if [[ -n "$API_KEY" ]]; then
            # Detect shell and add to appropriate config file
            SHELL_CONFIG=""
            if [[ "$SHELL" == *"zsh"* ]]; then
                SHELL_CONFIG="$HOME/.zshrc"
            elif [[ "$SHELL" == *"bash"* ]]; then
                SHELL_CONFIG="$HOME/.bashrc"
            else
                SHELL_CONFIG="$HOME/.profile"
            fi
            
            echo "" >> "$SHELL_CONFIG"
            echo "# Ask Terminal API Key" >> "$SHELL_CONFIG"
            echo "export GEMINI_API_KEY=\"$API_KEY\"" >> "$SHELL_CONFIG"
            print_success "API key added to $SHELL_CONFIG"
            print_warning "Please restart your terminal or run: source $SHELL_CONFIG"
        fi
    else
        echo
        print_warning "You can set up the API key later by adding this to your shell config:"
        echo "export GEMINI_API_KEY=\"your-api-key-here\""
    fi
}

# Create config directory
setup_config() {
    if [[ ! -d "$CONFIG_DIR" ]]; then
        mkdir -p "$CONFIG_DIR"
        print_success "Created config directory: $CONFIG_DIR"
        
        # Create default config
        cat > "$CONFIG_DIR/config.yaml" << EOF
default_mode: ai
show_explanations: true
auto_confirm: false
timeout: 30s
EOF
        print_success "Created default configuration"
    fi
}

# Add shell integration (optional)
setup_shell_integration() {
    echo
    read -p "Add shell alias 'ask' for easier access? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Detect shell config file
        SHELL_CONFIG=""
        if [[ "$SHELL" == *"zsh"* ]]; then
            SHELL_CONFIG="$HOME/.zshrc"
        elif [[ "$SHELL" == *"bash"* ]]; then
            SHELL_CONFIG="$HOME/.bashrc"
        else
            SHELL_CONFIG="$HOME/.profile"
        fi
        
        if ! grep -q "alias ask=" "$SHELL_CONFIG" 2>/dev/null; then
            echo "" >> "$SHELL_CONFIG"
            echo "# Ask Terminal alias" >> "$SHELL_CONFIG"
            echo "alias ask=\"$BINARY_NAME\"" >> "$SHELL_CONFIG"
            print_success "Added 'ask' alias to $SHELL_CONFIG"
            echo "You can now use: ask"
        else
            print_warning "Alias already exists in $SHELL_CONFIG"
        fi
    fi
}

# Verify installation
verify_installation() {
    if command -v $BINARY_NAME &> /dev/null; then
        print_success "Installation successful!"
        echo
        echo "🚀 Quick start:"
        echo "  $BINARY_NAME              # Launch Ask Terminal"
        echo "  $BINARY_NAME --help       # Show help"
        
        if command -v ask &> /dev/null 2>&1 || grep -q "alias ask=" "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" 2>/dev/null; then
            echo "  ask                       # Short alias"
        fi
        
        echo
        return 0
    else
        print_error "Installation verification failed"
        return 1
    fi
}

# Uninstall function
uninstall() {
    print_status "Uninstalling Ask Terminal..."
    
    # Remove binary
    if [[ -f "$INSTALL_DIR/$BINARY_NAME" ]]; then
        sudo rm "$INSTALL_DIR/$BINARY_NAME"
        print_success "Removed binary from $INSTALL_DIR"
    fi
    
    # Ask about config
    read -p "Remove configuration directory? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        print_success "Removed config directory"
    fi
    
    # Ask about shell config
    read -p "Remove API key and aliases from shell config? (y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        for config_file in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
            if [[ -f "$config_file" ]]; then
                # Remove API key line
                sed -i '/export GEMINI_API_KEY=/d' "$config_file" 2>/dev/null || true
                # Remove alias line
                sed -i '/alias ask=/d' "$config_file" 2>/dev/null || true
                # Remove comment lines
                sed -i '/# Ask Terminal/d' "$config_file" 2>/dev/null || true
            fi
        done
        print_success "Cleaned shell configuration files"
        print_warning "Please restart your terminal to apply changes"
    fi
    
    print_success "Uninstallation complete"
}

# Main installation flow
main() {
    echo "========================================"
    echo "     Ask Terminal Installer 🤖"
    echo "========================================"
    echo
    
    # Check for uninstall flag
    if [[ "$1" == "--uninstall" ]]; then
        uninstall
        exit 0
    fi
    
    # Check for help
    if [[ "$1" == "--help" || "$1" == "-h" ]]; then
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --uninstall    Uninstall Ask Terminal"
        echo "  --help, -h     Show this help message"
        echo
        echo "Examples:"
        echo "  curl -sSL https://raw.githubusercontent.com/jayaprakash-shanmugam/Ask-Terminal/main/install.sh | bash"
        echo "  curl -sSL https://raw.githubusercontent.com/jayaprakash-shanmugam/Ask-Terminal/main/install.sh | bash -s -- --uninstall"
        exit 0
    fi
    
    # Check if already installed
    if command -v $BINARY_NAME &> /dev/null; then
        print_warning "Ask Terminal is already installed"
        read -p "Reinstall? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation cancelled"
            exit 0
        fi
    fi
    
    # Installation steps
    check_go
    install_binary
    setup_config
    setup_api_key
    setup_shell_integration
    
    echo
    echo "========================================"
    verify_installation
    echo "========================================"
    
    echo
    echo "🎉 Installation complete!"
    echo
    echo "Next steps:"
    if [[ -z "$GEMINI_API_KEY" ]]; then
        echo "1. Get your free API key: https://makersuite.google.com/app/apikey"
        echo "2. Add to your shell: export GEMINI_API_KEY=\"your-key\""
        echo "3. Restart terminal or run: source ~/.bashrc"
    else
        echo "1. Restart your terminal or run: source ~/.bashrc"
    fi
    echo "2. Launch: $BINARY_NAME"
    echo "3. Try: 'show me all files' or 'help'"
    echo
    echo "Happy coding! 🚀"
}

# Run main function
main "$@"