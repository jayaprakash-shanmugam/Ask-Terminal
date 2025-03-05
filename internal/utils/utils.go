package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// const envFileName = ".env"

// EnsureAPIKey checks for the API key and prompts user to create one if not exists
func EnsureAPIKey() string {
	// First, try to read existing API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")

	// If API key exists in environment, return it
	if apiKey != "" {
		return apiKey
	}

	// Try to read from .env file
	// apiKey = readAPIKeyFromFile()
	// if apiKey != "" {
	// 	return apiKey
	// }

	// If no API key found, prompt user to enter
	apiKey = promptForAPIKey()
	// Set the environment variable
	os.Setenv("GEMINI_API_KEY", apiKey)

	// Verify it
	// fmt.Println("GEMINI_API_KEY set to:", os.Getenv("GEMINI_API_KEY"))

	// Write the API key to .env file
	// err := writeAPIKeyToFile(apiKey)
	// if err != nil {
	// 	log.Printf("Error writing API key to file: %v", err)
	// }

	return apiKey
}

// readAPIKeyFromFile reads the API key from .env file
// func readAPIKeyFromFile() string {
// 	file, err := os.Open(envFileName)
// 	if err != nil {
// 		return ""
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.HasPrefix(line, "GEMINI_API_KEY=") {
// 			return strings.TrimPrefix(line, "GEMINI_API_KEY=")
// 		}
// 	}

// 	return ""
// }

// promptForAPIKey asks user to input API key
func promptForAPIKey() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("No Gemini API key found.")
		fmt.Println("Please visit https://makersuite.google.com/app/apikey to create an API key")
		fmt.Print("Enter your Gemini API Key: ")

		apiKey, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		// Trim whitespace and newline
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			fmt.Println("API key cannot be empty. Please try again.")
			continue
		}

		return apiKey
	}
}

// writeAPIKeyToFile writes the API key to .env file
// func writeAPIKeyToFile(apiKey string) error {
// 	// Open file with write and create permissions
// 	file, err := os.Create(envFileName)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	// Write the API key to the file
// 	_, err = file.WriteString(fmt.Sprintf("GEMINI_API_KEY=%s\n", apiKey))
// 	return err
// }

// LoadAPIKey is a convenience method to get the API key
func LoadAPIKey() string {
	return EnsureAPIKey()
}
