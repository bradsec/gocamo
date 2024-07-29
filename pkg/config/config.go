package config

import (
	"flag"
	"fmt"
	"runtime"
)

type Config struct {
	Width         int
	Height        int
	BasePixelSize int
	JSONFile      string
	OutputDir     string
	ColorsString  string
	Cores         int
	AddEdge       bool
	AddNoise      bool
	PatternType   string
	ImageDir      string
	KValue        int
}

type CamoColors struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Width, "w", 1500, "Set the image width")
	flag.IntVar(&cfg.Height, "h", 1500, "Set the image height")
	flag.IntVar(&cfg.BasePixelSize, "b", 4, "Set the base pixel size (will be adjusted if necessary)")
	flag.StringVar(&cfg.JSONFile, "j", "", "Process a JSON file containing a list of color palettes")
	flag.StringVar(&cfg.OutputDir, "o", "output", "The output directory for generated images")
	flag.StringVar(&cfg.ColorsString, "c", "", "Generate a single pattern using a comma-separated list of hex colors")
	flag.IntVar(&cfg.Cores, "cores", runtime.NumCPU(), fmt.Sprintf("Number of CPU cores to use (1-%d available)", runtime.NumCPU()))
	flag.BoolVar(&cfg.AddEdge, "edge", false, "Add edge details to the pattern")
	flag.BoolVar(&cfg.AddNoise, "noise", false, "Add noise to the pattern")
	flag.StringVar(&cfg.PatternType, "t", "box", "Set the pattern type (blob, box, or image)")
	flag.StringVar(&cfg.ImageDir, "i", "input", "Input directory containing images for image-based camouflage")
	flag.IntVar(&cfg.KValue, "k", 4, "Number of main colors for image-based camouflage")

	flag.Parse()

	return cfg
}
