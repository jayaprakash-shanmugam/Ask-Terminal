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
}

func NewTerminalService(terminal *model.Terminal, gemini *GeminiService) *TerminalService {
	return &TerminalService{
		terminal: terminal,
		gemini:   gemini,
	}
}

// Run starts the terminal
func (t *TerminalService) Run() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Gemini Terminal - Control your computer with natural language")
	fmt.Println("Try commands like 'show all files', 'create a folder called data', or 'search for text files containing password'")
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println("Commands will be listed for your approval before execution")

	for t.terminal.IsRunning() {
		fmt.Printf("%s ", t.terminal.Prompt())

		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		t.terminal.AddToHistory(input)

		// Handle immediate exit command without calling API
		if input == "exit" || input == "quit" {
			t.terminal.SetRunning(false)
			fmt.Println("Goodbye!")
			continue
		}

		// Process with Gemini
		err = t.ProcessWithGemini(input)
		if err != nil {
			// COMMENTED
			// fmt.Printf("Error: %s\n", err)
		}
	}

	return nil
}

// ProcessWithGemini sends the query to Gemini and lists/confirms the returned command
func (t *TerminalService) ProcessWithGemini(query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmdResponse, err := t.gemini.GetCommandFromGemini(ctx, t.terminal.SystemPrompt(), query, t.terminal.APIKey())
	if err != nil {
		// COMMENTED
		// return fmt.Errorf("gemini error: %v", err)
		return nil
	}

	// Print the explanation
	fmt.Printf("\033[32m💡 %s\033[0m\n", cmdResponse.Explanation)

	// Format and display the command
	commandString := t.formatCommand(cmdResponse.Command, cmdResponse.Args)
	fmt.Printf("\033[33mCommand: %s\033[0m\n", commandString)

	// Ask for confirmation before execution
	return t.ConfirmAndExecute(cmdResponse)
}

// formatCommand formats a command and its arguments as a shell command string
func (t *TerminalService) formatCommand(command string, args []string) string {
	// Handle simple commands without args
	if len(args) == 0 {
		return command
	}

	// Format the command with arguments
	var formattedArgs []string
	for _, arg := range args {
		// Quote arguments containing spaces
		if strings.Contains(arg, " ") {
			formattedArgs = append(formattedArgs, fmt.Sprintf("\"%s\"", arg))
		} else {
			formattedArgs = append(formattedArgs, arg)
		}
	}

	return fmt.Sprintf("%s %s", command, strings.Join(formattedArgs, " "))
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
				// Could implement command history navigation if needed
				continue
			case 66: // Down arrow
				// Could implement command history navigation if needed
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

	// Parse the edited command
	parts := strings.Fields(editedCmd)
	if len(parts) == 0 {
		return nil
	}

	// Create edited command response
	editedCmdResponse := dto.CommandResponse{
		Command:     parts[0],
		Args:        parts[1:],
		Explanation: "User-edited command",
	}

	// Restore terminal to its original state before executing the command
	term.Restore(int(os.Stdin.Fd()), oldState)

	// Execute the edited command and capture output
	var outputBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	// Create a pipe to capture command output
	cmdOutput := func() error {
		cmd := exec.Command(editedCmdResponse.Command, editedCmdResponse.Args...)
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
		fmt.Println("Goodbye!")
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
	}

	// Handle relative paths
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(t.terminal.WorkingDir(), dir)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	t.terminal.SetWorkingDir(dir)
	return nil
}

// ShowHistory displays command history
func (t *TerminalService) ShowHistory() {
	for i, cmd := range t.terminal.History() {
		fmt.Printf("%d: %s\n", i+1, cmd)
	}
}

// ShowHelp displays help information
func (t *TerminalService) ShowHelp() {
	fmt.Println("Gemini Terminal Help:")
	fmt.Println("  Just type what you want to do in natural language!")
	fmt.Println("  For example:")
	fmt.Println("    - \"show me all text files\"")
	fmt.Println("    - \"create a folder called projects\"")
	fmt.Println("    - \"search for files modified in the last week\"")
	fmt.Println("    - \"compress all the images in this folder\"")
	fmt.Println()
	fmt.Println("  When a command is suggested:")
	fmt.Println("    - Press 'y' to execute the command")
	fmt.Println("    - Press 'n' to skip execution")
	fmt.Println("    - Press 'e' to edit the command before execution")
	fmt.Println()
	fmt.Println("  Some built-in commands:")
	fmt.Println("    - \"exit\" or \"quit\": Exit the terminal")
	fmt.Println("    - \"clear\": Clear the screen")
	fmt.Println("    - \"history\": Show command history")
	fmt.Println("    - \"help\": Display this help")
}
