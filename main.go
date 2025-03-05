package main

import (
	"fmt"
	"os"

	"askterminal/internal/model"
	"askterminal/internal/service"
)

func main() {
	// Create terminal model
	terminalModel := model.NewTerminal()

	// Create Gemini service
	geminiService := service.NewGeminiService()

	// Create terminal service
	terminalService := service.NewTerminalService(terminalModel, geminiService)

	// Run the terminal
	if err := terminalService.Run(); err != nil {
		fmt.Printf("Terminal error: %s\n", err)
		os.Exit(1)
	}
}
