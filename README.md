# Ask Terminal 🤖

An intelligent terminal that understands natural language commands and converts them to actual shell commands using Google's Gemini AI.

## ✨ Features

- **Natural Language Commands**: Type what you want in plain English
- **Smart Command Generation**: AI converts your requests to proper shell commands
- **Safety First**: Preview commands before execution
- **Dual Mode**: Switch between AI and traditional terminal modes
- **Context Aware**: Understands your current directory and files

## 🚀 Quick Start

### Option 1: One-line Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/jayaprakash-shanmugam/Ask-Terminal/main/install.sh | bash
```

### Option 2: Manual Install
```bash
# Clone and build
git clone https://github.com/jayaprakash-shanmugam/Ask-Terminal.git
cd Ask-Terminal
go build -o askterminal main.go
sudo mv askterminal /usr/local/bin/

# Setup API key
export GEMINI_API_KEY="your-gemini-api-key"
```

### Option 3: Go Install
```bash
go install github.com/jayaprakash-shanmugam/Ask-Terminal@latest
export GEMINI_API_KEY="your-gemini-api-key"
```

## 🔑 Get Your API Key

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key (it's free!)
3. Add to your shell profile:
```bash
echo 'export GEMINI_API_KEY="your-api-key-here"' >> ~/.bashrc
source ~/.bashrc
```

## 💻 Usage

### Launch
```bash
askterminal
```

### AI Mode Examples
```bash
🤖 show me all go files
🤖 create a folder called projects
🤖 find files containing "password" 
🤖 compress this directory
🤖 what files are in my home directory?
🤖 run the go program
```

### Switch Modes
```bash
mode          # Check current mode
askmode on       # Switch to AI mode  
askmode off    # Switch to traditional shell mode
help          # Show available commands
```

### Exit Options
```bash
exit          # Return to your default terminal
quit          # Same as exit
shell         # Launch your default shell temporarily
```

## 🛡️ Safety Features

- **Command Preview**: See exactly what will run before execution
- **User Confirmation**: Approve with `y`, skip with `n`, edit with `e`
- **Dangerous Command Protection**: Built-in safeguards against harmful operations
- **Context Awareness**: AI understands your current directory and files

## 📊 Command Examples

| What you say | Generated command |
|--------------|-------------------|
| `show all files` | `ls -la` |
| `find python files` | `find . -name "*.py"` |
| `go to home directory` | `cd ~` |
| `show disk usage` | `df -h` |
| `compress folder data` | `tar -czf data.tar.gz data/` |
| `search for TODO in code` | `grep -r "TODO" --include="*.go" .` |

## ⚙️ Configuration

### Environment Variables
```bash
export GEMINI_API_KEY="your-api-key"     # Required for AI features
export ASKTERMINAL_TIMEOUT="30s"         # API timeout (optional)
export ASKTERMINAL_MODE="ai"             # Default mode (optional)
```

### Config File (Optional)
Create `~/.config/askterminal/config.yaml`:
```yaml
default_mode: ai
show_explanations: true
auto_confirm: false
timeout: 30s
```

## 🔧 Troubleshooting

### API Key Issues
```bash
# Test your setup
askterminal --test

# Set API key interactively  
askterminal --setup
```

### Common Problems
- **"command not found"**: Make sure `/usr/local/bin` is in your `$PATH`
- **API timeout**: Check your internet connection and API key
- **Permission denied**: Run `chmod +x /usr/local/bin/askterminal`

## 🗑️ Uninstall

```bash
# Using install script
curl -sSL https://raw.githubusercontent.com/jayaprakash-shanmugam/Ask-Terminal/main/install.sh | bash -s -- --uninstall

# Manual removal
sudo rm /usr/local/bin/askterminal
rm -rf ~/.config/askterminal

# Remove from shell profile
# Edit ~/.bashrc or ~/.zshrc and remove GEMINI_API_KEY export
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

## 📝 Development

```bash
# Clone repository
git clone https://github.com/jayaprakash-shanmugam/Ask-Terminal.git
cd Ask-Terminal

# Install dependencies
go mod tidy

# Run in development
go run main.go

# Build
go build -o askterminal main.go
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Google Gemini AI for natural language processing
- The Go community for excellent tooling
- All contributors and users

---

**Made with ❤️ for developers who want to talk to their terminal**

## Support

- **Issues**: [GitHub Issues](https://github.com/jayaprakash-shanmugam/Ask-Terminal/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jayaprakash-shanmugam/Ask-Terminal/discussions)



# Newly Added Features
- **Multi-Pipeline Command Execution** - Multi Stage Commands are Executed in a normal instead of complex llm usage, Contributed by @NithishNithi