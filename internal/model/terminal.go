package model

import (
	"askterminal/internal/utils"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// Theme-adaptive colors that work on both light and dark backgrounds
const (
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"

	// Safe colors that work on both light and dark backgrounds
	brightBlue    = "\033[94m" // Bright blue - visible on both
	brightGreen   = "\033[92m" // Bright green - visible on both
	brightYellow  = "\033[93m" // Bright yellow - visible on both
	brightRed     = "\033[91m" // Bright red - visible on both
	brightMagenta = "\033[95m" // Bright magenta - visible on both
	brightCyan    = "\033[96m" // Bright cyan - visible on both

	// Neutral colors
	darkGray     = "\033[90m" // Dark gray - good for secondary info
	defaultColor = "\033[39m" // Terminal default color

	// Symbols that work universally
	symbolFolder = "📁"
	symbolFile   = "📄"
)

// Terminal represents our Gemini-enhanced terminal
type Terminal struct {
	prompt       string
	history      []string
	workingDir   string
	running      bool
	apiKey       string
	systemPrompt string
	config       *TerminalConfig
	askMode      bool
}

// Terminal configuration based on environment variables
type TerminalConfig struct {
	Style     string // enhanced
	ShowTime  bool
	ShowFiles bool
	ShowGit   bool
	ShowUser  bool
}

// Terminal status information
type TerminalStatus struct {
	Mode        string
	GitBranch   string
	GitStatus   string
	FileCount   int
	DirCount    int
	Username    string
	Hostname    string
	CurrentTime string
	LastCommand string
	CommandTime time.Duration
	PathDisplay string
}

// NewTerminal creates a new terminal instance
func NewTerminal() *Terminal {
	wd, _ := os.Getwd()

	// Load API key from environment
	apiKey := utils.LoadAPIKey()
	fmt.Println("API Key successfully loaded:", apiKey[:5]+"...")

	// System prompt that instructs Gemini how to interpret commands
	systemPrompt := utils.SystemPrompt

	// Load configuration from environment
	config := loadConfigFromEnv()

	return &Terminal{
		history:      []string{},
		workingDir:   wd,
		running:      true,
		apiKey:       apiKey,
		systemPrompt: systemPrompt,
		config:       config,
	}
}

// Load configuration from environment variables
func loadConfigFromEnv() *TerminalConfig {
	config := &TerminalConfig{
		ShowTime:  getEnvBool("ASK_TERMINAL_SHOW_TIME", true),
		ShowFiles: getEnvBool("ASK_TERMINAL_SHOW_FILES", true),
		ShowGit:   getEnvBool("ASK_TERMINAL_SHOW_GIT", true),
		ShowUser:  getEnvBool("ASK_TERMINAL_SHOW_USER", true),
	}

	return config
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

// Main prompt method that delegates based on configuration
func (t *Terminal) Prompt() string {
	return t.PromptWithMode("terminal")
}

func (t *Terminal) PromptWithMode(mode string) string {
	// Always update working directory
	wd, _ := os.Getwd()
	t.workingDir = wd

	return t.enhancedPrompt(mode)
}

// Enhanced prompt with full status information
func (t *Terminal) enhancedPrompt(mode string) string {
	status := t.gatherStatus(mode)

	var prompt strings.Builder

	// Main prompt line
	prompt.WriteString(t.buildMainPromptLine(status))

	return prompt.String()
}

func (t *Terminal) buildMainPromptLine(status *TerminalStatus) string {
	var line strings.Builder

	// Path with folder icon
	line.WriteString(fmt.Sprintf("%s%s%s",
		bold, status.PathDisplay, reset))

	// Git branch on same line (compact)
	if t.config.ShowGit && status.GitBranch != "" {
		gitColor := brightGreen
		if status.GitStatus != "clean" {
			gitColor = brightYellow
		}
		line.WriteString(fmt.Sprintf(" %s(%s)%s", gitColor, status.GitBranch, reset))
	}

	// Mode indicator
	var modeIndicator string
	if t.askMode {
		modeIndicator = fmt.Sprintf(" %smode:ask%s", brightMagenta, reset)
	} else {
		modeIndicator = fmt.Sprintf(" %smode:terminal%s", brightBlue, reset)
	}
	line.WriteString(modeIndicator)

	// Time
	if t.config.ShowTime {
		line.WriteString(fmt.Sprintf(" %s%s",
			status.CurrentTime, reset))
	}

	// User info
	if t.config.ShowUser && status.Username != "" {
		line.WriteString(fmt.Sprintf(" %s%s",
			status.Username, reset))
	}

	// New line with prompt symbol
	line.WriteString(fmt.Sprintf("\n❯%s ", reset))

	return line.String()
}

func (t *Terminal) gatherStatus(mode string) *TerminalStatus {
	wd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	displayPath := strings.Replace(wd, homeDir, "~", 1)

	status := &TerminalStatus{
		Mode:        mode,
		CurrentTime: time.Now().Format("15:04:05"),
		PathDisplay: displayPath,
	}

	// Get git info
	status.GitBranch = t.getCurrentBranch()
	status.GitStatus = t.getGitStatus()

	// Count files and directories
	status.FileCount, status.DirCount = t.countItems()

	// Get user info
	if user, err := user.Current(); err == nil {
		status.Username = user.Username
	}

	if hostname, err := os.Hostname(); err == nil {
		status.Hostname = hostname
	}

	return status
}

func (t *Terminal) getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

func (t *Terminal) getGitStatus() string {
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return ""
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return "clean"
	}

	lines := strings.Split(output, "\n")
	modified := 0
	untracked := 0
	staged := 0

	for _, line := range lines {
		if len(line) >= 2 {
			switch {
			case line[0] != ' ' && line[0] != '?' && line[0] != '!':
				staged++
			case line[1] == 'M':
				modified++
			case line[0] == '?' && line[1] == '?':
				untracked++
			}
		}
	}

	var parts []string
	if staged > 0 {
		parts = append(parts, fmt.Sprintf("+%d", staged))
	}
	if modified > 0 {
		parts = append(parts, fmt.Sprintf("~%d", modified))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("?%d", untracked))
	}

	if len(parts) == 0 {
		return "clean"
	}

	return strings.Join(parts, " ")
}

func (t *Terminal) countItems() (files, dirs int) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, 0
	}

	entries, err := os.ReadDir(wd)
	if err != nil {
		return 0, 0
	}

	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			dirs++
		} else {
			files++
		}
	}

	return files, dirs
}

func (t *Terminal) History() []string {
	return t.history
}

func (t *Terminal) WorkingDir() string {
	return t.workingDir
}

func (t *Terminal) IsRunning() bool {
	return t.running
}

func (t *Terminal) APIKey() string {
	return t.apiKey
}

func (t *Terminal) SystemPrompt() string {
	return t.systemPrompt
}

// Setters
func (t *Terminal) SetRunning(running bool) {
	t.running = running
}

func (t *Terminal) SetWorkingDir(dir string) {
	t.workingDir = dir
}

func (t *Terminal) AddToHistory(command string) {
	t.history = append(t.history, command)
}

// Configuration methods
func (t *Terminal) GetConfig() *TerminalConfig {
	return t.config
}

func (t *Terminal) SetPromptStyle(style string) {
	t.config.Style = style
}

func (t *Terminal) ToggleTime() {
	t.config.ShowTime = !t.config.ShowTime
}

func (t *Terminal) ToggleFiles() {
	t.config.ShowFiles = !t.config.ShowFiles
}

func (t *Terminal) ToggleGit() {
	t.config.ShowGit = !t.config.ShowGit
}

func (t *Terminal) ToggleUser() {
	t.config.ShowUser = !t.config.ShowUser
}

// Backward compatibility - remove the old GetCurrentBranch function
func GetCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

func (t *Terminal) SetToAskMode() {
	t.askMode = true
}

func (t *Terminal) SetToShellMode() {
	t.askMode = false
}
