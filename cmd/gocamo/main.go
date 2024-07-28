package main

import (
	"encoding/json"
	"fmt"
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

	if cfg.PatternType == "image" {
		imagePaths, err = utils.GetImageFiles(cfg.ImageDir)
		if err != nil {
			return fmt.Errorf("failed to get image files: %w", err)
		}
		if len(imagePaths) == 0 {
			return fmt.Errorf("no image files found in directory: %s", cfg.ImageDir)
		}
	} else {
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
		} else {
			return fmt.Errorf("no input specified. Use -c for colors, -j for JSON file, or -i for image directory")
		}
	}

	fmt.Printf("Generating patterns with dimensions %dx%d, base pixel size %d\n", cfg.Width, cfg.Height, cfg.BasePixelSize)
	if cfg.PatternType == "image" {
		fmt.Printf("Processing %d images using %d CPU cores\n", len(imagePaths), cfg.Cores)
	} else {
		fmt.Printf("Processing %d color palette(s) using %d CPU cores\n", len(camoList), cfg.Cores)
	}
	fmt.Printf("Pattern type: %s\n", cfg.PatternType)
	fmt.Printf("Add edge details: %v, Add noise: %v\n", cfg.AddEdge, cfg.AddNoise)
	fmt.Printf("Output path: %s\n\n", outputAbsPath)

	jobs := make(chan worker.Job, max(len(camoList), len(imagePaths)))
	results := make(chan error, max(len(camoList), len(imagePaths)))
	progressDone := make(chan bool)
	var wg sync.WaitGroup

	for w := 1; w <= cfg.Cores; w++ {
		wg.Add(1)
		go worker.Work(jobs, results, &wg)
	}

	totalJobs := max(len(camoList), len(imagePaths))
	go utils.TrackProgress(results, totalJobs, progressDone)

	if cfg.PatternType == "image" {
		imagePaths, err := utils.GetImageFiles(cfg.ImageDir)
		if err != nil {
			return fmt.Errorf("failed to get image files: %w", err)
		}
		if len(imagePaths) == 0 {
			return fmt.Errorf("no image files found in directory: %s", cfg.ImageDir)
		}
	}

	if cfg.PatternType == "image" {
		for i, imagePath := range imagePaths {
			jobs <- worker.Job{
				ImagePath:  imagePath,
				Index:      i,
				Config:     cfg,
				OutputPath: outputAbsPath,
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

	wg.Wait()
	close(results)

	<-progressDone

	duration := time.Since(startTime)
	fmt.Printf("\nRuntime %.2f seconds.\n", duration.Seconds())

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
