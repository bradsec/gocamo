package worker

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bradsec/gocamo/pkg/config"
)

func TestJob_Struct(t *testing.T) {
	camo := config.CamoColors{
		Name:   "test",
		Colors: []string{"ff0000", "00ff00"},
	}

	cfg := &config.Config{
		Width:         100,
		Height:        100,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	job := Job{
		Camo:       camo,
		ImagePath:  "/path/to/image.jpg",
		Index:      5,
		Config:     cfg,
		OutputPath: "/output/path",
	}

	// Test that all fields are accessible
	if job.Camo.Name != "test" {
		t.Errorf("Job.Camo.Name = %q, want %q", job.Camo.Name, "test")
	}
	if job.ImagePath != "/path/to/image.jpg" {
		t.Errorf("Job.ImagePath = %q, want %q", job.ImagePath, "/path/to/image.jpg")
	}
	if job.Index != 5 {
		t.Errorf("Job.Index = %d, want 5", job.Index)
	}
	if job.Config != cfg {
		t.Error("Job.Config should point to the same config instance")
	}
	if job.OutputPath != "/output/path" {
		t.Errorf("Job.OutputPath = %q, want %q", job.OutputPath, "/output/path")
	}
}

func TestWork_BoxPattern(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "box",
		AddEdge:       false,
		AddNoise:      false,
	}

	camo := config.CamoColors{
		Name:   "test_box",
		Colors: []string{"ff0000", "00ff00", "0000ff"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	// Create and send job
	job := Job{
		Camo:       camo,
		Index:      0,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	// Start worker
	wg.Add(1)
	go Work(jobs, results, &wg)

	// Wait for completion
	wg.Wait()
	close(results)

	// Check result
	err := <-results
	if err != nil {
		t.Errorf("Work() returned error: %v", err)
	}

	// Check that output file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}

	// Check filename format
	filename := files[0].Name()
	if !strings.Contains(filename, "gocamo_000_test_box_") {
		t.Errorf("Filename should contain pattern info: %s", filename)
	}
}

func TestWork_BlobPattern(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "blob",
		AddEdge:       true,
		AddNoise:      true,
	}

	camo := config.CamoColors{
		Name:   "test_blob",
		Colors: []string{"#ff0000", "#00ff00"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		Camo:       camo,
		Index:      1,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err := <-results
	if err != nil {
		t.Errorf("Work() returned error: %v", err)
	}

	// Check that output file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}
}

func TestWork_ImagePattern(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
			} else {
				img.Set(x, y, color.RGBA{0, 255, 0, 255}) // Green
			}
		}
	}

	tempDir := t.TempDir()
	testImagePath := filepath.Join(tempDir, "test_input.png")

	file, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	outputDir := filepath.Join(tempDir, "output")
	err = os.Mkdir(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	cfg := &config.Config{
		Width:         30,
		Height:        30,
		BasePixelSize: 4,
		PatternType:   "image",
		KValue:        3,
		AddEdge:       false,
		AddNoise:      false,
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		ImagePath:  testImagePath,
		Index:      0,
		Config:     cfg,
		OutputPath: outputDir,
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err = <-results
	if err != nil {
		t.Errorf("Work() returned error: %v", err)
	}

	// Check that output file was created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 output file, got %d", len(files))
	}

	// Check filename format for image-based pattern
	filename := files[0].Name()
	if !strings.Contains(filename, "gocamo_from_image_test_input_000_") {
		t.Errorf("Filename should contain image pattern info: %s", filename)
	}
	if !strings.Contains(filename, "_k3_") {
		t.Errorf("Filename should contain k-value: %s", filename)
	}
}

func TestWork_InvalidColorPattern(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	// Invalid colors
	camo := config.CamoColors{
		Name:   "invalid",
		Colors: []string{"invalid_color", "also_invalid"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		Camo:       camo,
		Index:      0,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err := <-results
	if err == nil {
		t.Error("Work() should return error for invalid colors")
	}
	if !strings.Contains(err.Error(), "hex") {
		t.Errorf("Expected error about hex colors, got: %v", err)
	}
}

func TestWork_InvalidImagePath(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "image",
		KValue:        4,
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		ImagePath:  "/nonexistent/image.jpg",
		Index:      0,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err := <-results
	if err == nil {
		t.Error("Work() should return error for nonexistent image")
	}
}

func TestWork_InvalidOutputPath(t *testing.T) {
	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	camo := config.CamoColors{
		Name:   "test",
		Colors: []string{"ff0000", "00ff00"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		Camo:       camo,
		Index:      0,
		Config:     cfg,
		OutputPath: "/invalid/output/path",
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err := <-results
	if err == nil {
		t.Error("Work() should return error for invalid output path")
	}
}

func TestWork_MultipleJobs(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         30,
		Height:        30,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	jobs := make(chan Job, 3)
	results := make(chan error, 3)
	var wg sync.WaitGroup

	// Create multiple jobs
	for i := 0; i < 3; i++ {
		camo := config.CamoColors{
			Name:   fmt.Sprintf("test_%d", i),
			Colors: []string{"ff0000", "00ff00"},
		}
		job := Job{
			Camo:       camo,
			Index:      i,
			Config:     cfg,
			OutputPath: tempDir,
		}
		jobs <- job
	}
	close(jobs)

	// Start worker
	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	// Check all results
	errorCount := 0
	for i := 0; i < 3; i++ {
		err := <-results
		if err != nil {
			errorCount++
			t.Errorf("Job %d failed: %v", i, err)
		}
	}

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	// Check that all output files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 output files, got %d", len(files))
	}
}

func TestWork_MultipleWorkers(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         25,
		Height:        25,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	jobCount := 6
	workerCount := 3

	jobs := make(chan Job, jobCount)
	results := make(chan error, jobCount)
	var wg sync.WaitGroup

	// Create jobs
	for i := 0; i < jobCount; i++ {
		camo := config.CamoColors{
			Name:   fmt.Sprintf("multi_%d", i),
			Colors: []string{"ff0000", "00ff00"},
		}
		job := Job{
			Camo:       camo,
			Index:      i,
			Config:     cfg,
			OutputPath: tempDir,
		}
		jobs <- job
	}
	close(jobs)

	// Start multiple workers
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go Work(jobs, results, &wg)
	}

	wg.Wait()
	close(results)

	// Check all results
	errorCount := 0
	for i := 0; i < jobCount; i++ {
		err := <-results
		if err != nil {
			errorCount++
			t.Errorf("Job failed: %v", err)
		}
	}

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}

	// Check that all output files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != jobCount {
		t.Errorf("Expected %d output files, got %d", jobCount, len(files))
	}
}

func TestWork_EmptyJobChannel(t *testing.T) {
	jobs := make(chan Job)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	// Close immediately (no jobs)
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	// Should complete quickly without error
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Expected behavior
	case <-time.After(1 * time.Second):
		t.Error("Work() should complete quickly with empty job channel")
	}
}

func TestWork_Timeout(t *testing.T) {
	// This test would be complex to implement reliably since it depends on timing
	// and the current implementation has a 60-second timeout.
	// We'll test the structure instead of actual timeout behavior.

	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         10, // Very small for quick processing
		Height:        10,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	camo := config.CamoColors{
		Name:   "timeout_test",
		Colors: []string{"ff0000", "00ff00"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		Camo:       camo,
		Index:      0,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	start := time.Now()
	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	duration := time.Since(start)
	close(results)

	err := <-results
	if err != nil {
		t.Errorf("Work() returned error: %v", err)
	}

	// Should complete much faster than 60 seconds
	if duration > 30*time.Second {
		t.Errorf("Work() took too long: %v", duration)
	}
}

func TestWork_MixedPatternTypes(t *testing.T) {
	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	tempDir := t.TempDir()
	testImagePath := filepath.Join(tempDir, "test.png")

	file, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	outputDir := filepath.Join(tempDir, "output")
	err = os.Mkdir(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	jobs := make(chan Job, 3)
	results := make(chan error, 3)
	var wg sync.WaitGroup

	// Box pattern job
	boxCfg := &config.Config{
		Width: 20, Height: 20, BasePixelSize: 4, PatternType: "box",
	}
	jobs <- Job{
		Camo:       config.CamoColors{Name: "box", Colors: []string{"ff0000", "00ff00"}},
		Index:      0,
		Config:     boxCfg,
		OutputPath: outputDir,
	}

	// Blob pattern job
	blobCfg := &config.Config{
		Width: 20, Height: 20, BasePixelSize: 4, PatternType: "blob",
	}
	jobs <- Job{
		Camo:       config.CamoColors{Name: "blob", Colors: []string{"0000ff", "ffff00"}},
		Index:      1,
		Config:     blobCfg,
		OutputPath: outputDir,
	}

	// Image pattern job
	imageCfg := &config.Config{
		Width: 20, Height: 20, BasePixelSize: 4, PatternType: "image", KValue: 2,
	}
	jobs <- Job{
		ImagePath:  testImagePath,
		Index:      2,
		Config:     imageCfg,
		OutputPath: outputDir,
	}

	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	// Check all results
	for i := 0; i < 3; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Mixed pattern job %d failed: %v", i, err)
		}
	}

	// Check that all files were created
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 output files, got %d", len(files))
	}
}

// Benchmark tests
func BenchmarkWork_BoxPattern(b *testing.B) {
	tempDir := b.TempDir()

	cfg := &config.Config{
		Width:         50,
		Height:        50,
		BasePixelSize: 4,
		PatternType:   "box",
		AddEdge:       false,
		AddNoise:      false,
	}

	camo := config.CamoColors{
		Name:   "bench",
		Colors: []string{"ff0000", "00ff00"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobs := make(chan Job, 1)
		results := make(chan error, 1)
		var wg sync.WaitGroup

		job := Job{
			Camo:       camo,
			Index:      i,
			Config:     cfg,
			OutputPath: tempDir,
		}
		jobs <- job
		close(jobs)

		wg.Add(1)
		go Work(jobs, results, &wg)
		wg.Wait()
		close(results)
		<-results // Consume result
	}
}

// Test edge cases
func TestWork_ZeroSizePattern(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         0, // Invalid size
		Height:        0,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	camo := config.CamoColors{
		Name:   "zero_size",
		Colors: []string{"ff0000", "00ff00"},
	}

	jobs := make(chan Job, 1)
	results := make(chan error, 1)
	var wg sync.WaitGroup

	job := Job{
		Camo:       camo,
		Index:      0,
		Config:     cfg,
		OutputPath: tempDir,
	}
	jobs <- job
	close(jobs)

	wg.Add(1)
	go Work(jobs, results, &wg)

	wg.Wait()
	close(results)

	err := <-results
	// Should handle zero size gracefully (may succeed with adjusted size or fail)
	if err != nil {
		t.Logf("Zero size pattern failed as expected: %v", err)
	}
}

// Test concurrent access to the same output directory
func TestWork_ConcurrentSameOutput(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Width:         20,
		Height:        20,
		BasePixelSize: 4,
		PatternType:   "box",
	}

	jobs := make(chan Job, 5)
	results := make(chan error, 5)
	var wg sync.WaitGroup

	// Create jobs that write to the same directory
	for i := 0; i < 5; i++ {
		camo := config.CamoColors{
			Name:   fmt.Sprintf("concurrent_%d", i),
			Colors: []string{"ff0000", "00ff00"},
		}
		job := Job{
			Camo:       camo,
			Index:      i,
			Config:     cfg,
			OutputPath: tempDir, // Same output directory
		}
		jobs <- job
	}
	close(jobs)

	// Start multiple workers
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go Work(jobs, results, &wg)
	}

	wg.Wait()
	close(results)

	// All should succeed without conflicts
	for i := 0; i < 5; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent job %d failed: %v", i, err)
		}
	}

	// Check that all files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("Expected 5 output files, got %d", len(files))
	}
}
