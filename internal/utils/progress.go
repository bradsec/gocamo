package utils

import (
	"fmt"
	"strings"
)

// PrintBanner displays the gocamo ASCII art banner to the console.
// This function prints a stylized text banner using block characters to provide
// visual branding for the application startup.
func PrintBanner() {
	banner := `
▒▀▀▀ ▒▀▀█ ▒▀▀▀ ▒▀▀▓ █▀▓▀█ ▒▀▀█
▓ ▀█ ▓  ▒ ▓    ▓▀▀▓ ▓ █ ▒ ▓  ▒
▀▀▀▀ ▀▀▀▀ ▀▀▀▀ ▀  ▀ ▀ ▀ ▀ ▀▀▀▀`
	fmt.Println(banner)
}

// TrackProgress monitors and displays a real-time progress bar for job completion.
// It receives job results through a channel and updates a visual progress bar showing
// completion percentage and counts. The function tracks both successful completions
// and errors, displaying error statistics when all jobs are finished.
//
// The progress bar uses Unicode block characters to create a visual representation
// of completion status and is updated in real-time as results are received.
//
// Parameters:
//   - results: A receive-only channel that delivers job completion results (nil for success, error for failure)
//   - total: The total number of jobs expected to complete
//   - done: A send-only channel used to signal when all jobs have completed
//
// The function will send true to the done channel and return when all jobs have been processed.
// If any jobs failed, it prints an error summary showing the number of failed jobs.
func TrackProgress(results <-chan error, total int, done chan<- bool) {
	// Always signal completion, even when total is zero or the results
	// channel closes early; otherwise the caller blocks forever.
	defer func() { done <- true }()

	completed := 0
	errors := 0
	for result := range results {
		if result != nil {
			errors++
		}
		completed++
		printProgressBar(completed, total, 50)
		if completed == total {
			break
		}
	}
	fmt.Println() // Print a newline after the progress bar
	if errors > 0 {
		fmt.Printf("%d out of %d jobs failed.\n", errors, total)
	}
}

func printProgressBar(done, total, width int) {
	if total <= 0 {
		return
	}

	percent := float64(done) / float64(total)
	filled := int(percent * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r[%s] %.1f%% (%d/%d)", bar, percent*100, done, total)
}
