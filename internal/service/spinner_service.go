package service

import (
	"time"

	"github.com/briandowns/spinner"
)

// SpinnerManager handles spinner operations
type SpinnerManager struct {
	spinner *spinner.Spinner
}

// NewSpinnerManager creates a new spinner manager
func NewSpinnerManager() *SpinnerManager {
	// Custom animation frames
	frames := []string{"▴▴", "▸▸", "▾▾", "◂◂"}

	s := spinner.New(frames, 500*time.Millisecond)
	s.Suffix = " Processing your request..."

	return &SpinnerManager{
		spinner: s,
	}
}

// Start begins the spinner
func (sm *SpinnerManager) Start() {
	sm.spinner.Start()
}

// Stop ends the spinner
func (sm *SpinnerManager) Stop() {
	sm.spinner.Stop()
}

// UpdateMessage changes the spinner message
func (sm *SpinnerManager) UpdateMessage(msg string) {
	sm.spinner.Suffix = " " + msg
}
