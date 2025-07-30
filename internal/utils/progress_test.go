package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPrintBanner(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintBanner()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that output contains expected banner elements
	if !strings.Contains(output, "▒") {
		t.Error("Banner should contain block characters")
	}
	if !strings.Contains(output, "▓") {
		t.Error("Banner should contain different block characters")
	}
	if !strings.Contains(output, "▀") {
		t.Error("Banner should contain top block characters")
	}

	// Check that it's multi-line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Errorf("Banner should have at least 3 lines, got %d", len(lines))
	}

	// Check that lines are not empty
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			t.Errorf("Banner line %d should not be empty", i)
		}
	}
}

func TestTrackProgress_AllSuccess(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create channels
	results := make(chan error, 3)
	done := make(chan bool, 1)

	// Start progress tracking in a goroutine
	go TrackProgress(results, 3, done)

	// Send successful results
	results <- nil
	results <- nil
	results <- nil
	close(results)

	// Wait for completion
	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that progress was displayed
	if !strings.Contains(output, "█") {
		t.Error("Progress should contain filled progress characters")
	}
	if !strings.Contains(output, "100.0%") {
		t.Error("Progress should show 100% completion")
	}
	if !strings.Contains(output, "(3/3)") {
		t.Error("Progress should show completed count")
	}
}

func TestTrackProgress_WithErrors(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create channels
	results := make(chan error, 3)
	done := make(chan bool, 1)

	// Start progress tracking in a goroutine
	go TrackProgress(results, 3, done)

	// Send mixed results
	results <- nil
	results <- fmt.Errorf("test error")
	results <- nil
	close(results)

	// Wait for completion
	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that error message was displayed
	if !strings.Contains(output, "1 out of 3 jobs failed") {
		t.Error("Progress should report failed jobs")
	}
}

func TestTrackProgress_AllErrors(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create channels
	results := make(chan error, 2)
	done := make(chan bool, 1)

	// Start progress tracking in a goroutine
	go TrackProgress(results, 2, done)

	// Send error results
	results <- fmt.Errorf("error 1")
	results <- fmt.Errorf("error 2")
	close(results)

	// Wait for completion
	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that all errors were reported
	if !strings.Contains(output, "2 out of 2 jobs failed") {
		t.Error("Progress should report all failed jobs")
	}
}

func TestTrackProgress_SingleJob(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create channels
	results := make(chan error, 1)
	done := make(chan bool, 1)

	// Start progress tracking in a goroutine
	go TrackProgress(results, 1, done)

	// Send single result
	results <- nil
	close(results)

	// Wait for completion
	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that progress was displayed correctly for single job
	if !strings.Contains(output, "100.0%") {
		t.Error("Progress should show 100% for single job")
	}
	if !strings.Contains(output, "(1/1)") {
		t.Error("Progress should show (1/1) for single job")
	}
}

func TestTrackProgress_NoJobs(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Create channels
	results := make(chan error)
	done := make(chan bool, 1)

	// Start progress tracking in a goroutine
	go TrackProgress(results, 0, done)

	// Close immediately (no jobs)
	close(results)

	// Wait for completion with timeout
	select {
	case <-done:
		// Should not complete since no jobs reach total
	case <-time.After(100 * time.Millisecond):
		// This is expected - progress tracker shouldn't complete with 0 total
	}

	w.Close()
	os.Stdout = old
}

func TestPrintProgressBar(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test various progress states
	tests := []struct {
		done, total, width int
		expectedPercent    float64
	}{
		{0, 10, 20, 0.0},
		{5, 10, 20, 50.0},
		{10, 10, 20, 100.0},
		{3, 4, 10, 75.0},
		{1, 3, 15, 33.3},
	}

	for _, tt := range tests {
		printProgressBar(tt.done, tt.total, tt.width)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that progress bars contain expected elements
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Progress bar should contain brackets")
	}
	if !strings.Contains(output, "█") {
		t.Error("Progress bar should contain filled characters")
	}
	if !strings.Contains(output, "░") {
		t.Error("Progress bar should contain empty characters")
	}
	if !strings.Contains(output, "%") {
		t.Error("Progress bar should contain percentage")
	}

	// Check specific percentages appear
	if !strings.Contains(output, "0.0%") {
		t.Error("Should show 0.0% for zero progress")
	}
	if !strings.Contains(output, "100.0%") {
		t.Error("Should show 100.0% for complete progress")
	}
	if !strings.Contains(output, "50.0%") {
		t.Error("Should show 50.0% for half progress")
	}
}

func TestPrintProgressBar_EdgeCases(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test edge cases
	printProgressBar(0, 1, 1)     // Minimum width
	printProgressBar(1, 1, 1)     // Complete with minimum width
	printProgressBar(0, 1, 100)   // Large width
	printProgressBar(50, 100, 50) // Exact half

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not crash and should contain basic elements
	if !strings.Contains(output, "[") {
		t.Error("Progress bar should contain opening bracket")
	}
	if !strings.Contains(output, "]") {
		t.Error("Progress bar should contain closing bracket")
	}
	if !strings.Contains(output, "%") {
		t.Error("Progress bar should contain percentage sign")
	}
}

func TestTrackProgress_Concurrent(t *testing.T) {
	// Test that TrackProgress can handle concurrent writes to results channel
	results := make(chan error, 100)
	done := make(chan bool, 1)

	// Capture stdout to avoid test output pollution
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	go TrackProgress(results, 100, done)

	var wg sync.WaitGroup
	// Send results from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if (start+j)%3 == 0 {
					results <- fmt.Errorf("error %d", start+j)
				} else {
					results <- nil
				}
			}
		}(i * 10)
	}

	wg.Wait()
	close(results)
	<-done

	w.Close()
	os.Stdout = old

	// Consume the output to prevent blocking
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Test passed if no panic occurred
}

func TestTrackProgress_LargeNumbers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large numbers test in short mode")
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := make(chan error, 1000)
	done := make(chan bool, 1)

	go TrackProgress(results, 1000, done)

	// Send many results quickly
	go func() {
		for i := 0; i < 1000; i++ {
			if i%100 == 0 {
				results <- fmt.Errorf("error %d", i)
			} else {
				results <- nil
			}
		}
		close(results)
	}()

	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should handle large numbers correctly
	if !strings.Contains(output, "100.0%") {
		t.Error("Should reach 100% with large numbers")
	}
	if !strings.Contains(output, "(1000/1000)") {
		t.Error("Should show correct completion count")
	}
	if !strings.Contains(output, "10 out of 1000 jobs failed") {
		t.Error("Should report correct error count")
	}
}

func TestTrackProgress_MixedErrorTypes(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := make(chan error, 5)
	done := make(chan bool, 1)

	go TrackProgress(results, 5, done)

	// Send different types of errors
	results <- nil
	results <- fmt.Errorf("io error")
	results <- fmt.Errorf("network timeout")
	results <- nil
	results <- fmt.Errorf("validation error")
	close(results)

	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should count all errors regardless of type
	if !strings.Contains(output, "3 out of 5 jobs failed") {
		t.Error("Should count all error types")
	}
}

// Benchmark tests
func BenchmarkPrintBanner(b *testing.B) {
	// Redirect stdout to discard output during benchmark
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	for i := 0; i < b.N; i++ {
		PrintBanner()
	}
}

func BenchmarkPrintProgressBar(b *testing.B) {
	// Redirect stdout to discard output during benchmark
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	for i := 0; i < b.N; i++ {
		printProgressBar(i%100, 100, 50)
	}
}

func BenchmarkTrackProgress(b *testing.B) {
	// Redirect stdout to discard output during benchmark
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	for i := 0; i < b.N; i++ {
		results := make(chan error, 10)
		done := make(chan bool, 1)

		go TrackProgress(results, 10, done)

		for j := 0; j < 10; j++ {
			results <- nil
		}
		close(results)
		<-done
	}
}

// Test that progress tracking handles rapid updates
func TestTrackProgress_RapidUpdates(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	results := make(chan error, 50)
	done := make(chan bool, 1)

	go TrackProgress(results, 50, done)

	// Send results very quickly
	go func() {
		for i := 0; i < 50; i++ {
			results <- nil
		}
		close(results)
	}()

	<-done

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should still reach completion
	if !strings.Contains(output, "100.0%") {
		t.Error("Should reach 100% with rapid updates")
	}
}

// Test progress bar width handling
func TestPrintProgressBar_WidthValidation(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with various widths
	widths := []int{1, 5, 10, 25, 50, 100}
	for _, width := range widths {
		printProgressBar(width/2, width, width)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should handle all widths without crashing
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(widths) {
		t.Errorf("Expected %d progress bars, got %d", len(widths), len(lines))
	}

	// Each line should contain a valid progress bar
	for i, line := range lines {
		if !strings.Contains(line, "[") || !strings.Contains(line, "]") {
			t.Errorf("Progress bar %d should contain brackets: %s", i, line)
		}
	}
}
