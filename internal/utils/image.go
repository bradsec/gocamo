package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LoadImage loads an image file from the filesystem and decodes it into an image.Image.
// It supports JPEG (.jpg, .jpeg) and PNG (.png) image formats. The function automatically
// detects the image format based on the file extension and uses the appropriate decoder.
//
// Parameters:
//   - filename: Path to the image file to load
//
// Returns the decoded image and an error if the file cannot be opened, the format is
// unsupported, or decoding fails.
//
// Supported formats:
//   - JPEG: .jpg, .jpeg extensions
//   - PNG: .png extension
//
// Example:
//
//	img, err := LoadImage("photo.jpg")
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get the file extension
	ext := strings.ToLower(filepath.Ext(filename))

	var img image.Image

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	return img, nil
}

// SaveImage encodes and saves an image to the provided writer in PNG format.
// The function uses PNG encoding to preserve image quality and transparency information.
//
// Parameters:
//   - img: The image to be encoded and saved
//   - w: The writer where the encoded PNG data will be written
//
// Returns an error if the PNG encoding fails or writing to the destination fails.
//
// Example:
//
//	file, err := os.Create("output.png")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//	err = SaveImage(img, file)
func SaveImage(img image.Image, w io.Writer) error {
	return png.Encode(w, img)
}

// GetImageFiles recursively scans a directory and returns a slice of paths to all image files found.
// It walks through the entire directory tree and identifies image files based on their extensions.
// Only files with supported image extensions are included in the results.
//
// Parameters:
//   - dir: The root directory path to scan for image files
//
// Returns a slice of file paths to all discovered image files and an error if the directory
// cannot be accessed or traversed.
//
// Supported image extensions:
//   - .jpg, .jpeg (JPEG format)
//   - .png (PNG format)
//
// Example:
//
//	imageFiles, err := GetImageFiles("./photos")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d image files\n", len(imageFiles))
func GetImageFiles(dir string) ([]string, error) {
	var images []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			images = append(images, path)
		}
		return nil
	})
	return images, err
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}
