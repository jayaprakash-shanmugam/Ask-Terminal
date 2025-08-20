package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"askterminal/internal/model"
	"askterminal/internal/service"
)

func main() {
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Create terminal model
	terminalModel := model.NewTerminal()

	// Create Gemini service
	geminiService := service.NewGeminiService()

	// Create terminal service
	terminalService := service.NewTerminalService(terminalModel, geminiService)

	// Handle graceful shutdown
	go func() {
		<-c
		fmt.Println("\nShutting down gracefully...")
		terminalModel.SetRunning(false)
		os.Exit(0)
	}()

	// Run the terminal
	if err := terminalService.Run(); err != nil {
		fmt.Printf("Terminal error: %s\n", err)
		os.Exit(1)
	}
}
