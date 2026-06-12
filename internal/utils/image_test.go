package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"JPEG file", "test.jpg", true},
		{"JPEG file uppercase", "test.JPG", true},
		{"JPEG file alt extension", "test.jpeg", true},
		{"JPEG file alt extension uppercase", "test.JPEG", true},
		{"PNG file", "test.png", true},
		{"PNG file uppercase", "test.PNG", true},
		{"Text file", "test.txt", false},
		{"PDF file", "test.pdf", false},
		{"No extension", "test", false},
		{"Hidden file", ".hidden", false},
		{"GIF file", "test.gif", false},
		{"BMP file", "test.bmp", false},
		{"TIFF file", "test.tiff", false},
		{"WebP file", "test.webp", false},
		{"Mixed case", "test.JpG", true},
		{"Multiple dots", "test.backup.jpg", true},
		{"Long path", "/very/long/path/to/image.png", true},
		{"Empty string", "", false},
		{"Just extension", ".jpg", true},
		{"Space in name", "test image.jpg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImageFile(tt.path)
			if result != tt.expected {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestLoadImage_FileNotFound(t *testing.T) {
	_, err := LoadImage("/nonexistent/file.jpg")
	if err == nil {
		t.Error("LoadImage should return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "error opening file") {
		t.Errorf("Expected error about opening file, got: %v", err)
	}
}

func TestLoadImage_UnsupportedFormat(t *testing.T) {
	// Create a temporary file with unsupported extension
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.gif")
	err := os.WriteFile(tempFile, []byte("fake gif content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = LoadImage(tempFile)
	if err == nil {
		t.Error("LoadImage should return error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported image format") {
		t.Errorf("Expected error about unsupported format, got: %v", err)
	}
}

func TestLoadImage_ValidPNG(t *testing.T) {
	// Create a valid PNG image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill with red color
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.png")

	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}
	file.Close()

	// Test loading the image
	loadedImg, err := LoadImage(tempFile)
	if err != nil {
		t.Errorf("LoadImage failed: %v", err)
	}

	bounds := loadedImg.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Loaded image size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_ValidJPEG(t *testing.T) {
	// Create a valid JPEG image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Fill with blue color
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
	}

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.jpg")

	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	if err != nil {
		t.Fatalf("Failed to encode JPEG: %v", err)
	}
	file.Close()

	// Test loading the image
	loadedImg, err := LoadImage(tempFile)
	if err != nil {
		t.Errorf("LoadImage failed: %v", err)
	}

	bounds := loadedImg.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Loaded image size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_InvalidImageData(t *testing.T) {
	// Create a file with .png extension but invalid content
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.png")
	err := os.WriteFile(tempFile, []byte("invalid png data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = LoadImage(tempFile)
	if err == nil {
		t.Error("LoadImage should return error for invalid PNG data")
	}
	if !strings.Contains(err.Error(), "error decoding image") {
		t.Errorf("Expected error about decoding image, got: %v", err)
	}
}

func TestSaveImage(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 50), uint8(y * 50), 100, 255})
		}
	}

	// Test saving to buffer
	var buf bytes.Buffer
	err := SaveImage(img, &buf)
	if err != nil {
		t.Errorf("SaveImage failed: %v", err)
	}

	// Verify the buffer contains PNG data
	if buf.Len() == 0 {
		t.Error("SaveImage wrote no data")
	}

	// Check PNG signature
	data := buf.Bytes()
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(data) < len(pngSignature) {
		t.Error("Saved data too short to contain PNG signature")
	} else {
		for i, b := range pngSignature {
			if data[i] != b {
				t.Error("Saved data does not start with PNG signature")
				break
			}
		}
	}
}

func TestGetImageFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	files, err := GetImageFiles(tempDir)
	if err != nil {
		t.Errorf("GetImageFiles failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}
}

func TestGetImageFiles_NonexistentDirectory(t *testing.T) {
	_, err := GetImageFiles("/nonexistent/directory")
	if err == nil {
		t.Error("GetImageFiles should return error for nonexistent directory")
	}
}

func TestGetImageFiles_WithImageFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"image1.jpg",
		"image2.png",
		"image3.jpeg",
		"document.txt", // Not an image
		"data.pdf",     // Not an image
		"photo.JPG",    // Uppercase
		"graphic.PNG",  // Uppercase
	}

	// Create the files
	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filepath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	files, err := GetImageFiles(tempDir)
	if err != nil {
		t.Errorf("GetImageFiles failed: %v", err)
	}

	// Should find 5 image files (jpg, png, jpeg, JPG, PNG)
	expectedCount := 5
	if len(files) != expectedCount {
		t.Errorf("Expected %d image files, got %d", expectedCount, len(files))
	}

	// Check that all returned files are image files
	for _, file := range files {
		if !isImageFile(file) {
			t.Errorf("GetImageFiles returned non-image file: %s", file)
		}
	}

	// Check that specific files are included
	expectedFiles := map[string]bool{
		"image1.jpg":  false,
		"image2.png":  false,
		"image3.jpeg": false,
		"photo.JPG":   false,
		"graphic.PNG": false,
	}

	for _, file := range files {
		basename := filepath.Base(file)
		if _, exists := expectedFiles[basename]; exists {
			expectedFiles[basename] = true
		}
	}

	for filename, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file %s not found in results", filename)
		}
	}
}

func TestGetImageFiles_WithSubdirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory structure
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create files in both directories
	mainFile := filepath.Join(tempDir, "main.jpg")
	subFile := filepath.Join(subDir, "sub.png")

	err = os.WriteFile(mainFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	err = os.WriteFile(subFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	files, err := GetImageFiles(tempDir)
	if err != nil {
		t.Errorf("GetImageFiles failed: %v", err)
	}

	// Should find both files (walks subdirectories)
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	// Check both files are found
	foundMain, foundSub := false, false
	for _, file := range files {
		if strings.HasSuffix(file, "main.jpg") {
			foundMain = true
		}
		if strings.HasSuffix(file, "sub.png") {
			foundSub = true
		}
	}

	if !foundMain {
		t.Error("main.jpg not found")
	}
	if !foundSub {
		t.Error("sub.png not found")
	}
}

func TestGetImageFiles_WithHiddenFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create normal and hidden image files
	normalFile := filepath.Join(tempDir, "normal.jpg")
	hiddenFile := filepath.Join(tempDir, ".hidden.jpg")

	err := os.WriteFile(normalFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create normal file: %v", err)
	}

	err = os.WriteFile(hiddenFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	files, err := GetImageFiles(tempDir)
	if err != nil {
		t.Errorf("GetImageFiles failed: %v", err)
	}

	// Should find both files (hidden files are still valid)
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestSaveImage_Integration(t *testing.T) {
	// Create an image
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}

	// Save to temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "saved.png")

	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	err = SaveImage(img, file)
	if err != nil {
		t.Errorf("SaveImage failed: %v", err)
	}
	file.Close()

	// Load it back and verify
	loadedImg, err := LoadImage(tempFile)
	if err != nil {
		t.Errorf("Failed to load saved image: %v", err)
	}

	bounds := loadedImg.Bounds()
	if bounds.Dx() != 3 || bounds.Dy() != 3 {
		t.Errorf("Loaded image size = %dx%d, want 3x3", bounds.Dx(), bounds.Dy())
	}
}

// Benchmark tests
func BenchmarkIsImageFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isImageFile("test.jpg")
	}
}

func BenchmarkLoadImage_Small(b *testing.B) {
	// Create a small test image file once
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "bench.png")

	file, err := os.Create(tempFile)
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadImage(tempFile)
	}
}

func BenchmarkSaveImage_Small(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		SaveImage(img, &buf)
	}
}

// Edge case testing
func TestLoadImage_EdgeCases(t *testing.T) {
	// Test loading files with unusual but valid extensions
	tempDir := t.TempDir()

	// Create image with .jpeg extension
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	jpegFile := filepath.Join(tempDir, "test.jpeg")

	file, err := os.Create(jpegFile)
	if err != nil {
		t.Fatalf("Failed to create JPEG file: %v", err)
	}
	jpeg.Encode(file, img, nil)
	file.Close()

	_, err = LoadImage(jpegFile)
	if err != nil {
		t.Errorf("Failed to load .jpeg file: %v", err)
	}
}

func TestGetImageFiles_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test as root user")
	}

	tempDir := t.TempDir()

	// Create a subdirectory with no read permissions
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	// GetImageFiles should handle permission errors gracefully
	_, err = GetImageFiles(tempDir)
	if err == nil {
		t.Error("Expected GetImageFiles to return error for permission denied")
	}
}

// Test concurrent access
func TestLoadImage_Concurrent(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "concurrent.png")

	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	// Load the same image concurrently
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := LoadImage(tempFile)
			if err != nil {
				t.Errorf("Concurrent LoadImage failed: %v", err)
			}
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
