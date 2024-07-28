package utils

import (
	"fmt"
	"strings"
)

func PrintBanner() {
	banner := `
▒▀▀▀ ▒▀▀█ ▒▀▀▀ ▒▀▀▓ █▀▓▀█ ▒▀▀█
▓ ▀█ ▓  ▒ ▓    ▓▀▀▓ ▓ █ ▒ ▓  ▒
▀▀▀▀ ▀▀▀▀ ▀▀▀▀ ▀  ▀ ▀ ▀ ▀ ▀▀▀▀`
	fmt.Println(banner)
}

func TrackProgress(results <-chan error, total int, done chan<- bool) {
	completed := 0
	errors := 0
	for result := range results {
		if result != nil {
			errors++
		}
		completed++
		printProgressBar(completed, total, 50)
		if completed == total {
			fmt.Println() // Print a newline after the progress bar is complete
			done <- true
			return
		}
	}
	if errors > 0 {
		fmt.Printf("\n%d out of %d jobs failed.\n", errors, total)
	} else {
		fmt.Println()
	}
}

func printProgressBar(done, total, width int) {
	percent := float64(done) / float64(total)
	filled := int(percent * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r[%s] %.1f%% (%d/%d)", bar, percent*100, done, total)
}
