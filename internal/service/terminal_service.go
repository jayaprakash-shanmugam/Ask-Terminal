package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"askterminal/internal/dto"
	"askterminal/internal/model"

	"golang.org/x/term"
)

type TerminalService struct {
	terminal *model.Terminal
	gemini   *GeminiService
	askMode  bool // New field to track ask mode
}

func NewTerminalService(terminal *model.Terminal, gemini *GeminiService) *TerminalService {
	t := terminal
	t.SetToAskMode()
	return &TerminalService{
		terminal: t,
		gemini:   gemini,
		askMode:  true, // Default to ask mode
	}
}

// Run starts the terminal
func (t *TerminalService) Run() error {
	reader := bufio.NewReader(os.Stdin)
	t.showWelcome()

	for t.terminal.IsRunning() {
		// Show different prompts based on mode
		if t.askMode {
			fmt.Printf("\n %s ", t.terminal.Prompt())
			t.terminal.SetToAskMode()
		} else {
			fmt.Printf("\n %s ", t.terminal.Prompt())
			t.terminal.SetToShellMode()
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		t.terminal.AddToHistory(input)

		// Handle mode switching and special commands
		if t.handleSpecialCommands(input) {
			continue
		}

		// Handle immediate exit command without calling API
		if input == "exit" || input == "quit" {
			t.terminal.SetRunning(false)
			terminationMsg := "Session terminated."
			if t.askMode {
				terminationMsg = "Alright, see you next time!"
			}
			fmt.Println(terminationMsg)
			continue
		}

		// Process based on current mode
		if t.askMode {
			err = t.ProcessWithGemini(input)
		} else {
			err = t.ProcessDirectCommand(input)
		}

		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	return nil
}

func (t *TerminalService) showWelcome() {
	fmt.Println("🚀 Enhanced Terminal - AI + Traditional Mode")
	fmt.Println("════════════════════════════════════════════")
	fmt.Println("🤖 Ask Mode: ON  - Use natural language commands")
	fmt.Println("💻 Terminal Mode: Available with 'askmode off'")
	fmt.Println()
	fmt.Println("Try commands like:")
	fmt.Println("  • 'show all files'")
	fmt.Println("  • 'create a folder called data'")
	fmt.Println("  • 'summarize current directory'")
	fmt.Println("  • 'find all go files'")
	fmt.Println("  • 'askmode off' to switch to traditional terminal")
	fmt.Println()
	fmt.Println("Type 'help' for more options or 'exit' to quit")
	fmt.Println("════════════════════════════════════════════")
}

func (t *TerminalService) handleSpecialCommands(input string) bool {
	switch strings.ToLower(input) {
	case "askmode on", "ask on", "ai on":
		t.askMode = true
		t.terminal.SetToAskMode()
		fmt.Println("🤖 Ask Mode: ON - You can now use natural language commands")
		return true
	case "askmode off", "ask off", "ai off":
		t.askMode = false
		t.terminal.SetToShellMode()
		fmt.Println("💻 Terminal Mode: ON - Using traditional terminal commands")
		return true
	case "mode":
		if t.askMode {
			fmt.Println("🤖 Current Mode: Ask Mode (AI-assisted)")
		} else {
			fmt.Println("💻 Current Mode: Terminal Mode (traditional)")
		}
		return true
	case "help":
		t.ShowHelp()
		return true
	case "summarize", "summary":
		if t.askMode {
			t.summarizeDirectory()
		} else {
			fmt.Println("Use 'askmode on' to enable AI features like summarize")
		}
		return true
	}
	return false
}

// ProcessDirectCommand handles traditional terminal commands
func (t *TerminalService) ProcessDirectCommand(input string) error {
	// Parse command and arguments
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	args := parts[1:]

	// Create CommandResponse for consistency
	cmdResponse := dto.CommandResponse{
		Command:       command,
		Args:          args,
		Explanation:   fmt.Sprintf("Executing: %s", input),
		SkipExecution: false,
	}

	// Execute directly without confirmation in terminal mode
	return t.ExecuteCommand(cmdResponse)
}

// ProcessWithGemini sends the query to Gemini and lists/confirms the returned command
func (t *TerminalService) ProcessWithGemini(query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Increased timeout
	defer cancel()

	dirInfo := t.CollectDirectoryContext()
	systemPrompt := fmt.Sprintf(
		"%s\n\nThe current directory contains the following files and folders:\n%s",
		t.terminal.SystemPrompt(), dirInfo,
	)

	// Handle special AI queries
	if t.isAnalysisQuery(query) {
		return t.handleAnalysisQuery(query, dirInfo)
	}

	cmdResponse, err := t.gemini.GetCommandFromGemini(ctx, systemPrompt, query, t.terminal.APIKey())
	if err != nil {
		return fmt.Errorf("gemini error: %v", err)
	}

	if cmdResponse.MultiExec {
		fmt.Println("🔄 Multiple commands detected. Executing step-by-step:")

		for idx, multiCmd := range cmdResponse.MultiCommands {
			// Show step header
			fmt.Printf("\n[Step %d/%d] %s\n", idx+1, len(cmdResponse.MultiCommands), multiCmd.Explanation)

			// Format and show the command
			commandString := t.formatCommand(multiCmd.Command, multiCmd.Args)
			fmt.Printf("\033[33mCommand: %s\033[0m\n", commandString)

			// Prepare a CommandResponse for execution
			commandRes := dto.CommandResponse{
				Command:       multiCmd.Command,
				Args:          multiCmd.Args,
				Explanation:   multiCmd.Explanation,
				SkipExecution: multiCmd.SkipExecution,
			}

			// If SkipExecution is true, just display without executing
			if multiCmd.SkipExecution {
				fmt.Println("⚠️ This step is marked to skip execution.")
				continue
			}

			// Confirm and execute the command
			if err := t.ConfirmAndExecute(commandRes); err != nil {
				fmt.Printf("❌ Error executing command: %s\n", err)
				return err
			}
			fmt.Println("✅ Step completed successfully")
		}
		return nil
	}

	// Print the explanation
	fmt.Printf("\033[32m💡 %s\033[0m\n", cmdResponse.Explanation)

	if cmdResponse.SkipExecution {
		fmt.Println("💬", strings.Join(cmdResponse.Args, " "))
		return nil
	}

	// Format and display the command
	commandString := t.formatCommand(cmdResponse.Command, cmdResponse.Args)
	fmt.Printf("\033[33mCommand: %s\033[0m\n", commandString)

	// Ask for confirmation before execution
	return t.ConfirmAndExecute(cmdResponse)
}

func (t *TerminalService) isAnalysisQuery(query string) bool {
	analysisKeywords := []string{
		"summarize", "summary", "analyze", "analysis", "describe", "overview",
		"what files", "what folders", "content", "structure", "explain",
	}

	lowerQuery := strings.ToLower(query)
	for _, keyword := range analysisKeywords {
		if strings.Contains(lowerQuery, keyword) {
			return true
		}
	}
	return false
}

func (t *TerminalService) handleAnalysisQuery(query, dirInfo string) error {
	fmt.Printf("\033[32m💡 Analyzing directory structure and contents\033[0m\n")

	// Enhanced directory analysis
	analysis := t.analyzeDirectoryStructure(dirInfo)
	fmt.Println(analysis)

	return nil
}

func (t *TerminalService) analyzeDirectoryStructure(dirInfo string) string {
	var analysis strings.Builder

	lines := strings.Split(strings.TrimSpace(dirInfo), "\n")
	if len(lines) == 0 {
		return "📁 Directory appears to be empty"
	}

	fileCount := 0
	dirCount := 0
	fileTypes := make(map[string]int)

	analysis.WriteString("📊 Directory Analysis:\n")
	analysis.WriteString("═══════════════════════\n\n")

	// Count items and categorize
	for _, line := range lines {
		if strings.HasPrefix(line, "[DIR]") {
			dirCount++
		} else if strings.HasPrefix(line, "[FILE]") {
			fileCount++
			// Extract file extension
			parts := strings.Fields(line)
			if len(parts) > 1 {
				filename := parts[1]
				ext := filepath.Ext(filename)
				if ext != "" {
					fileTypes[ext]++
				} else {
					fileTypes["no extension"]++
				}
			}
		}
	}

	analysis.WriteString(fmt.Sprintf("📂 Directories: %d\n", dirCount))
	analysis.WriteString(fmt.Sprintf("📄 Files: %d\n\n", fileCount))

	if len(fileTypes) > 0 {
		analysis.WriteString("📋 File Types:\n")
		for ext, count := range fileTypes {
			if ext == "no extension" {
				analysis.WriteString(fmt.Sprintf("  • No extension: %d\n", count))
			} else {
				analysis.WriteString(fmt.Sprintf("  • %s files: %d\n", ext, count))
			}
		}
		analysis.WriteString("\n")
	}

	// Show structure
	analysis.WriteString("🗂️  Directory Structure:\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "[DIR]") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				analysis.WriteString(fmt.Sprintf("  📁 %s/\n", parts[1]))
			}
		}
	}

	analysis.WriteString("\n📄 Files:\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "[FILE]") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				filename := parts[1]
				ext := filepath.Ext(filename)
				icon := t.getFileIcon(ext)
				analysis.WriteString(fmt.Sprintf("  %s %s\n", icon, filename))
			}
		}
	}

	return analysis.String()
}

func (t *TerminalService) getFileIcon(ext string) string {
	switch ext {
	case ".go":
		return "🐹"
	case ".js", ".ts":
		return "📜"
	case ".html", ".htm":
		return "🌐"
	case ".css":
		return "🎨"
	case ".md":
		return "📖"
	case ".json":
		return "⚙️"
	case ".txt":
		return "📝"
	case ".pdf":
		return "📄"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "🖼️"
	case ".mp4", ".avi", ".mov":
		return "🎥"
	case ".mp3", ".wav":
		return "🎵"
	case ".zip", ".tar", ".gz":
		return "📦"
	default:
		return "📄"
	}
}

func (t *TerminalService) summarizeDirectory() {
	dirInfo := t.CollectDirectoryContext()
	analysis := t.analyzeDirectoryStructure(dirInfo)
	fmt.Println(analysis)
}

func (t *TerminalService) CollectDirectoryContext() string {
	var builder strings.Builder

	err := filepath.Walk(t.terminal.WorkingDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignore errors
		}
		relPath, _ := filepath.Rel(t.terminal.WorkingDir(), path)
		if relPath == "." {
			return nil
		}

		// Skip hidden files and common ignore patterns
		if strings.HasPrefix(info.Name(), ".") ||
			info.Name() == "node_modules" ||
			info.Name() == "vendor" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			builder.WriteString(fmt.Sprintf("[DIR]  %s\n", relPath))
		} else {
			builder.WriteString(fmt.Sprintf("[FILE] %s\n", relPath))
		}
		return nil
	})

	if err != nil {
		return "Unable to read directory"
	}

	return builder.String()
}

// ConfirmAndExecute asks for user confirmation before executing a command
func (t *TerminalService) ConfirmAndExecute(cmdResponse dto.CommandResponse) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Execute this command? [y/n/e]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "y", "yes":
		// Execute the command
		return t.ExecuteCommand(cmdResponse)
	case "e", "edit":
		// Allow user to edit the command
		return t.EditAndExecuteCommand(cmdResponse)
	default:
		fmt.Println("Command not executed.")
		return nil
	}
}

func (t *TerminalService) EditAndExecuteCommand(cmdResponse dto.CommandResponse) error {
	// Prepare the initial command
	currentCmd := t.formatCommand(cmdResponse.Command, cmdResponse.Args)

	// Get the original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Create a buffer for editing
	cmdBuffer := []rune(currentCmd)
	cursorPos := len(cmdBuffer)

	// Redraw the entire line
	redrawLine := func() {
		// Clear the current line
		fmt.Print("\r\033[K")
		fmt.Print("Edit command: ", string(cmdBuffer))

		// Move cursor back to the correct position
		fmt.Printf("\r\033[%dC", len("Edit command: ")+cursorPos)
	}

	// Highlight output with ANSI color codes
	highlightOutput := func(output []byte, isError bool) string {
		// Choose color based on output type
		var colorCode string
		if isError {
			colorCode = "\033[31m" // Red for errors
		} else {
			colorCode = "\033[32m" // Green for standard output
		}

		// Reset color code
		resetCode := "\033[0m"

		// Apply coloring
		return colorCode + string(output) + resetCode
	}

	// Show response in a separate section
	showResponse := func(stdout, stderr []byte) {
		fmt.Println("\n--- Command Output ---")
		fmt.Println("")

		// Display standard output (if any)
		if len(stdout) > 0 {
			fmt.Println(highlightOutput(stdout, false))
		}

		// Display error output (if any)
		if len(stderr) > 0 {
			fmt.Println(highlightOutput(stderr, true))
		}
	}

	// Initial draw
	redrawLine()

	// Read and process input
	for {
		var b [3]byte
		os.Stdin.Read(b[:])

		// Handle special keys
		if b[0] == 27 && b[1] == 91 {
			switch b[2] {
			case 65: // Up arrow
				continue
			case 66: // Down arrow
				continue
			case 67: // Right arrow
				if cursorPos < len(cmdBuffer) {
					cursorPos++
					redrawLine()
				}
				continue
			case 68: // Left arrow
				if cursorPos > 0 {
					cursorPos--
					redrawLine()
				}
				continue
			}
		}

		// Handle enter key
		if b[0] == 13 {
			fmt.Println() // Move to next line
			break
		}

		// Handle backspace
		if b[0] == 127 {
			if cursorPos > 0 {
				cmdBuffer = append(cmdBuffer[:cursorPos-1], cmdBuffer[cursorPos:]...)
				cursorPos--
				redrawLine()
			}
			continue
		}

		// Handle printable characters
		if b[0] >= 32 && b[0] < 127 {
			// Insert character at cursor position
			cmdBuffer = append(cmdBuffer[:cursorPos], append([]rune{rune(b[0])}, cmdBuffer[cursorPos:]...)...)
			cursorPos++
			redrawLine()
		}
	}

	// Convert buffer to string and trim
	editedCmd := strings.TrimSpace(string(cmdBuffer))

	// Handle empty input
	if editedCmd == "" {
		fmt.Println("Command not executed.")
		return nil
	}

	// Restore terminal to its original state before executing the command
	term.Restore(int(os.Stdin.Fd()), oldState)

	// Execute the edited command and capture output
	var outputBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	// Create a pipe to capture command output
	cmdOutput := func() error {
		// Use bash to support shell operators like &&
		cmd := exec.Command("bash", "-c", editedCmd)
		cmd.Stdout = &outputBuffer
		cmd.Stderr = &errBuffer

		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}

	// Execute the command and handle potential errors
	err = cmdOutput()

	// Prepare response buffer with both stdout and stderr
	stdout := outputBuffer.Bytes()
	stderr := errBuffer.Bytes()

	// Show the command output with syntax highlighting
	showResponse(stdout, stderr)

	return err
}

// formatCommand combines command and args into a single string
func (t *TerminalService) formatCommand(cmd string, args []string) string {
	return cmd + " " + strings.Join(args, " ")
}

// ExecuteCommand runs the command returned by Gemini
func (t *TerminalService) ExecuteCommand(cmdResponse dto.CommandResponse) error {
	command := cmdResponse.Command
	args := cmdResponse.Args

	// Handle built-in commands
	switch command {
	case "cd":
		return t.ChangeDirectory(args)
	case "pwd":
		fmt.Println(t.terminal.WorkingDir())
		return nil
	case "history":
		t.ShowHistory()
		return nil
	case "clear":
		fmt.Print("\033[H\033[2J")
		return nil
	case "help":
		t.ShowHelp()
		return nil
	case "exit", "quit":
		t.terminal.SetRunning(false)
		terminationMsg := "Session terminated."
		if t.askMode {
			terminationMsg = "Alright, see you next time!"
		}
		fmt.Println(terminationMsg)
		return nil
	}

	// Execute external command
	cmd := exec.Command(command, args...)
	cmd.Dir = t.terminal.WorkingDir()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (t *TerminalService) ChangeDirectory(args []string) error {
	var dir string

	if len(args) < 1 {
		// If no arguments provided, go to home directory
		dir = os.Getenv("HOME")
	} else {
		dir = args[0]

		// Manually expand ~ to HOME
		if strings.HasPrefix(dir, "~") {
			home := os.Getenv("HOME")
			dir = filepath.Join(home, strings.TrimPrefix(dir, "~"))
		}
	}

	// Handle relative paths
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(t.terminal.WorkingDir(), dir)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}

	t.terminal.SetWorkingDir(dir)
	return nil
}

// ShowHistory displays command history
func (t *TerminalService) ShowHistory() {
	fmt.Println("📜 Command History:")
	fmt.Println("═══════════════════")
	for i, cmd := range t.terminal.History() {
		fmt.Printf("%3d: %s\n", i+1, cmd)
	}
}

// ShowHelp displays help information
func (t *TerminalService) ShowHelp() {
	fmt.Println("🚀 Enhanced Terminal Help")
	fmt.Println("═════════════════════════")
	fmt.Println()
	fmt.Println("🤖 Ask Mode Commands (AI-assisted):")
	fmt.Println("  • 'show me all text files'")
	fmt.Println("  • 'create a folder called projects'")
	fmt.Println("  • 'search for files modified in the last week'")
	fmt.Println("  • 'summarize current directory'")
	fmt.Println("  • 'find all go files'")
	fmt.Println("  • 'is there any file called main.go?'")
	fmt.Println("  • 'run that go file'")
	fmt.Println()
	fmt.Println("💻 Terminal Mode Commands (traditional):")
	fmt.Println("  • ls, dir - list files")
	fmt.Println("  • cd <directory> - change directory")
	fmt.Println("  • pwd - show current directory")
	fmt.Println("  • mkdir <name> - create directory")
	fmt.Println("  • touch <name> - create file")
	fmt.Println("  • cat <file> - view file content")
	fmt.Println()
	fmt.Println("🔄 Mode Switching:")
	fmt.Println("  • 'askmode on' or 'ai on' - Enable AI mode")
	fmt.Println("  • 'askmode off' or 'ai off' - Enable terminal mode")
	fmt.Println("  • 'mode' - Show current mode")
	fmt.Println()
	fmt.Println("🛠️  Special Commands:")
	fmt.Println("  • 'summarize' - Analyze current directory")
	fmt.Println("  • 'history' - Show command history")
	fmt.Println("  • 'clear' - Clear screen")
	fmt.Println("  • 'help' - Show this help")
	fmt.Println("  • 'exit' or 'quit' - Exit terminal")
	fmt.Println()
	fmt.Println("⚡ Command Execution:")
	fmt.Println("  • Press 'y' to execute suggested command")
	fmt.Println("  • Press 'n' to skip execution")
	fmt.Println("  • Press 'e' to edit command before execution")
}
