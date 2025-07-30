# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
```bash
# Build the application
go build -o gocamo ./cmd/gocamo

# Run tests for all packages  
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/generator
```

### Development
```bash
# Generate a single pattern with colors
./gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat1 -w 900 -h 900

# Process multiple color palettes from JSON
./gocamo -j colors.json -t pat2

# Generate image-based patterns
./gocamo -t image -b 10

# Generate all pattern types for comparison
./gocamo -c "#ffffff,#012169,#e4002b" -t all
```

## Architecture

### Core Components

**Main Application Flow (cmd/gocamo/main.go)**
- Parses CLI flags via `config.ParseFlags()`
- Sets up concurrent worker pool with goroutines
- Distributes pattern generation jobs across CPU cores
- Supports both color palette and image-based pattern generation

**Pattern Generation (internal/generator/)**
- `Generator` interface implemented by 4 pattern types:
  - `Pat1Generator`: Military-inspired layered patterns
  - `Pat2Generator`: MultiCam-inspired woodland patterns  
  - `Pat3Generator`: Digital angular/geometric patterns
  - `Pat4Generator`: Organic blob patterns
- `ImageGenerator`: Extracts colors from input images using k-means clustering
- All generators use context for timeout control and cancellation

**Worker Pool (internal/worker/)**
- Concurrent job processing with 60-second timeouts per task
- Jobs contain either color palette (`CamoColors`) or image path
- Results channel collects completion status from all workers

**Configuration (pkg/config/)**
- CLI flag parsing with validation for colors, dimensions, cores
- Supports JSON file input for batch processing multiple color palettes
- Color ratio customization (equal, random, or custom ratios)
- Pattern type auto-detection based on input flags

### Key Design Patterns

**Concurrency**: Worker pool pattern with configurable core count for performance on multi-core systems

**Interface-based Generation**: All pattern types implement the `Generator` interface for consistent API

**Context-based Control**: All generation operations support cancellation and timeouts

**Modular Structure**: Clear separation between CLI, configuration, generation logic, and worker coordination

### File Organization
- `cmd/gocamo/`: Main application entry point
- `internal/generator/`: Pattern generation algorithms (pat1.go, pat2.go, etc.)
- `internal/utils/`: Shared utilities for colors, images, and progress tracking  
- `internal/worker/`: Concurrent job processing
- `pkg/config/`: Configuration parsing and validation
- `input/`: Default directory for source images when using `-t image`
- `output/`: Default directory for generated pattern images

### Pattern Types
- **pat1**: Traditional military camouflage with organic and digital elements
- **pat2**: MultiCam-inspired with color blending and twig patterns
- **pat3**: Sharp geometric digital camouflage with cellular automata
- **pat4**: Organic blob patterns using flowing cellular automata rules
- **image**: Analyzes input photos to extract dominant colors for pattern generation
- **all**: Generates all 4 pattern types for each color palette