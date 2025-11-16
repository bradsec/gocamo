package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bradsec/gocamo/internal/utils"
	"github.com/bradsec/gocamo/internal/worker"
	"github.com/bradsec/gocamo/pkg/config"
)

func main() {
	// Seed random number generator for unpredictable patterns
	rand.Seed(time.Now().UnixNano())

	cfg := config.ParseFlags()

	utils.PrintBanner()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	startTime := time.Now()

	outputAbsPath, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.MkdirAll(outputAbsPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var camoList []config.CamoColors
	var imagePaths []string

	if cfg.ColorsString != "" && strings.TrimSpace(cfg.ColorsString) == "" {
		return fmt.Errorf("no valid colors provided in color string")
	}

	// Handle input type and validation
	switch cfg.PatternType {
	case "image":
		imagePaths, err = utils.GetImageFiles(cfg.ImageDir)
		if err != nil {
			return fmt.Errorf("failed to get image files: %w", err)
		}
		if len(imagePaths) == 0 {
			return fmt.Errorf("no image files found in directory: %s", cfg.ImageDir)
		}
	case "pat1", "pat2", "pat3", "pat4", "pat5":
		if cfg.ColorsString != "" {
			colors := strings.Split(cfg.ColorsString, ",")
			camoList = append(camoList, config.CamoColors{Name: "custom", Colors: colors})
		} else if cfg.JSONFile != "" {
			file, err := os.Open(cfg.JSONFile)
			if err != nil {
				return fmt.Errorf("failed to open JSON file: %w", err)
			}
			defer file.Close()

			if err := json.NewDecoder(file).Decode(&camoList); err != nil {
				return fmt.Errorf("failed to decode JSON: %w", err)
			}
			if len(camoList) == 0 {
				return fmt.Errorf("no color palettes found in JSON file")
			}
		} else {
			return fmt.Errorf("no input specified. Use -c for colors, -j for JSON file, or -i for image directory")
		}
	case "all":
		if cfg.ColorsString != "" {
			colors := strings.Split(cfg.ColorsString, ",")
			camoList = append(camoList, config.CamoColors{Name: "custom", Colors: colors})
		} else if cfg.JSONFile != "" {
			file, err := os.Open(cfg.JSONFile)
			if err != nil {
				return fmt.Errorf("failed to open JSON file: %w", err)
			}
			defer file.Close()

			if err := json.NewDecoder(file).Decode(&camoList); err != nil {
				return fmt.Errorf("failed to decode JSON: %w", err)
			}
			if len(camoList) == 0 {
				return fmt.Errorf("no color palettes found in JSON file")
			}
		} else {
			return fmt.Errorf("no input specified. Use -c for colors, -j for JSON file, or -i for image directory")
		}
	default:
		return fmt.Errorf("invalid pattern type: %s (must be 'pat1', 'pat2', 'pat3', 'pat4', 'pat5', 'all', or 'image')", cfg.PatternType)
	}

	// Print configuration information
	fmt.Printf("Generating patterns with dimensions %dx%d, base pixel size %d\n", cfg.Width, cfg.Height, cfg.BasePixelSize)
	if cfg.PatternType == "image" {
		fmt.Printf("Processing %d images using %d CPU cores\n", len(imagePaths), cfg.Cores)
	} else {
		fmt.Printf("Processing %d color palette(s) using %d CPU cores\n", len(camoList), cfg.Cores)
	}
	fmt.Printf("Pattern type: %s\n", cfg.PatternType)
	fmt.Printf("Add edge details: %v, Add noise: %v\n", cfg.AddEdge, cfg.AddNoise)
	fmt.Printf("Output path: %s\n\n", outputAbsPath)

	// Set up worker pools and channels
	var totalJobs int
	if cfg.PatternType == "all" {
		totalJobs = len(camoList) * 5 // 5 pattern types for each color palette
	} else {
		totalJobs = max(len(camoList), len(imagePaths))
	}

	jobs := make(chan worker.Job, totalJobs)
	results := make(chan error, totalJobs)
	progressDone := make(chan bool)
	var wg sync.WaitGroup

	// Start worker pool
	for w := 1; w <= cfg.Cores; w++ {
		wg.Add(1)
		go worker.Work(jobs, results, &wg)
	}

	// Start progress tracking
	go utils.TrackProgress(results, totalJobs, progressDone)

	// Queue jobs based on pattern type
	if cfg.PatternType == "image" {
		for i, imagePath := range imagePaths {
			jobs <- worker.Job{
				ImagePath:  imagePath,
				Index:      i,
				Config:     cfg,
				OutputPath: outputAbsPath,
			}
		}
	} else if cfg.PatternType == "all" {
		// Generate all 5 pattern types for each color palette
		patternTypes := []string{"pat1", "pat2", "pat3", "pat4", "pat5"}
		jobIndex := 0
		for _, camo := range camoList {
			for _, patType := range patternTypes {
				// Create a copy of the config with the specific pattern type
				configCopy := *cfg
				configCopy.PatternType = patType

				jobs <- worker.Job{
					Camo:       camo,
					Index:      jobIndex,
					Config:     &configCopy,
					OutputPath: outputAbsPath,
				}
				jobIndex++
			}
		}
	} else {
		for i, camo := range camoList {
			jobs <- worker.Job{
				Camo:       camo,
				Index:      i,
				Config:     cfg,
				OutputPath: outputAbsPath,
			}
		}
	}
	close(jobs)

	// Wait for all jobs to complete
	wg.Wait()
	close(results)
	<-progressDone

	duration := time.Since(startTime)
	fmt.Printf("\nRuntime %.2f seconds.\n", duration.Seconds())

	return nil
}
