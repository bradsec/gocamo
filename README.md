# GOCAMO

## A digital camouflage pattern image generator written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/bradsec/gocamo)](https://goreportcard.com/report/github.com/bradsec/gocamo)

GOCAMO is a Go program that generates military-styled digital camouflage patterns. The patterns can be generated using custom color palettes specified in a JSON file or via command-line arguments. Images are saved in PNG format in the specified `output` directory. The output filename shows the HEX colors used and the resolution of the image. Two or more colors can be used in pattern palettes for patterns pat1, pat2, pat3, pat4, pat5.

## Features

- Generate digital camouflage patterns with customizable colors, unique patterns, and any resolution
- Configurable base pixel size for different pattern granularity
- Output images include color codes in the filename for easy reference
- Multi-core processing for improved performance when generating multiple patterns

## Generation Speed

Generation speed depends on the number of images, resolution, and base pixel size. Higher resolution and smaller base pixel sizes require more processing time. The program uses Go's concurrency features to leverage multiple CPU cores when processing multiple color palettes from a JSON file, significantly improving performance on multi-core systems.

## Optimized File Size

The program will produce optimized small PNG file sizes for high-resolution patterns (when generating without `-noise` or `-edge`):
- 280kB for a 4K image (`-w 3840 -h 2160`)
- 1.6MB for a 4K image with `-noise` added
- 9.4MB for a 4K image with `-edge` details added
- 10.5MB for a 4K image with `-noise` and `-edge` details added

## Pattern Types (pat1, pat2, pat3, pat4, all, image)

### pat1 (set using `-t pat1`, default if no type specified)

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat1 -w 900 -h 900
```

![Sample Images](samples/gocamo_000_custom_46482f_6d6851_9b967f_1e2415_pat1_w900x900.png)

### pat2 (set using `-t pat2`)

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat2 -w 900 -h 900
```

![Sample Images](samples/gocamo_001_custom_46482f_6d6851_9b967f_1e2415_pat2_w900x900.png)

### pat3 (set using `-t pat3`)

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat3 -w 900 -h 900
```

![Sample Images](samples/gocamo_002_custom_46482f_6d6851_9b967f_1e2415_pat3_w900x900.png)

### pat4 (set using `-t pat4`)

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat4 -w 900 -h 900
```

![Sample Images](samples/gocamo_003_custom_46482f_6d6851_9b967f_1e2415_pat4_w900x900.png)

### pat5 (set using `-t pat5`)

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t pat5 -w 900 -h 900
```

![Sample Images](samples/gocamo_004_custom_46482f_6d6851_9b967f_1e2415_pat5_w900x900.png)

### all (set using `-t all`)
The all option generates patterns using all five pattern types (pat1, pat2, pat3, pat4, pat5) for each color palette provided. This is useful when you want to see all pattern variations for comparison or when generating a complete set of patterns from the same color scheme.

```terminal
gocamo -c "#46482f,#6d6851,#9b967f,#1e2415" -t all -w 900 -h 900
```

When using `-t all`, the tool will generate 5 patterns for each color palette:
- One pattern using pat1 algorithm
- One pattern using pat2 algorithm  
- One pattern using pat3 algorithm
- One pattern using pat4 algorithm
- One pattern using pat5 algorithm

### image (set using `-t image`, uses images in the `input` directory as reference)
The ImageGenerator processes an input image to create a camouflage-like pattern based on the original image's colors and features. Loads the input image and resizes it to the target dimensions while maintaining aspect ratio. Applies max pooling to reduce the image size and enhance prominent features. Applies a Laplacian filter to enhance edges and details in the image. Uses k-means clustering to extract the main colors from the processed image. Maps each pixel in the processed image to the closest main color.

Reference (source) photo:

![Sample Images](input/photo_jungle.jpg)

```terminal
gocamo -t image -w 900 -h 900
```

Pattern result with default `-b 4 and k 4`:

![Sample Images](samples/image.png)

```terminal
gocamo -t image -w 900 -h 900 -b 10
```

Pattern result with `-b 10`:

![Sample Images](samples/image_b10.png)


```terminal
gocamo -t image -w 900 -h 900 -k 16
```

Pattern result with default `-b 4` and 16 colors `k 16`:

![Sample Images](samples/imagek16.png)

```terminal
gocamo -t image -w 900 -h 900 -b 10 -k 16
```

Pattern result with `-b 10` and 16 colors `k 16`:

![Sample Images](samples/imageb10k16.png)




## Installing

### Option 1 Download the pre-built Binary files from [Releases](https://github.com/bradsec/gocamo/releases)

### Option 2 Use Go to install the latest version
If you have Go installed (https://go.dev/doc/install) you can install the latest version of gocamo with this command:
```terminal
go install github.com/bradsec/gocamo/cmd/gocamo@latest
````

### Option 3 Clone Repo and Build

```terminal
git clone https://github.com/bradsec/gocamo.git
cd gocamo
go build -o gocamo ./cmd/gocamo
# Copy the gocamo binary/executable to a directory in your system PATH
```

## Examples commands

1. Generate a single pattern with specified colors (defaults to pat1 pattern):
   ```
   # Three color pat1 pattern (default)
   gocamo -c "#ffffff,#012169,#e4002b"

   # Three color pat4 pattern
   gocamo -c "#ffffff,#012169,#e4002b" -t pat4

   # Three color pat3 pattern
   gocamo -c "#ffffff,#012169,#e4002b" -t pat3

   # Generate all 4 pattern types with same colors
   gocamo -c "#ffffff,#012169,#e4002b" -t all
   ```

2. Process multiple color palettes from a JSON file:
   ```
   # All color schemes as pat4 patterns
   gocamo -j colors.json -t pat4

   # Process all color schemes in `colors.json` (defaults to pat1)
   gocamo -j colors.json

   # Generate all 5 pattern types for each color scheme in JSON file
   gocamo -j colors.json -t all
   ```

3. Make pattern from images use `-t image`, this option looks in the image input directory default `input` and processes the images, identifying clusters of colors to produce patterns based on the images. Will batch process any images in the directory. Change input directory with `-i` flag. Use `-b` to increase block pixel size in output pattern.
   ```
   gocamo -t image -b 10
   ```

3. Set custom dimensions:
   ```
   gocamo -j colors.json -w 3840 -h 2160
   ```

4. Set base pixel size (increase of decease pixels in patterns):
   ```
   gocamo -c "#ffffff,#012169,#e4002b" -b 6
   ```

5. Specify output directory:
   ```
   gocamo -j colors.json -o output_folder
   ```

6. Use specific number of CPU cores:
   ```
   gocamo -j colors.json -cores 4
   ```
7. Add noise to image (larger file size)
   ```
   gocamo -c "#ffffff,#012169,#e4002b" -noise
   ```
8. Add edge details to image (larger file size)
   ```
   gocamo -c "#ffffff,#012169,#e4002b" -edge
   ```
9. Add noise and edge details (largest file size)
   ```
   gocamo -c "#ffffff,#012169,#e4002b" -noise -edge
   ```

## Paths

- New patterns will save to output directory (default is output)

## Command Line Usage

```
Usage of ./gocamo:
  -b int
    	Set the base pixel size (will be adjusted if necessary) (default 4)
  -c string
    	Generate a single pattern using a comma-separated list of hex colors
  -cores int
    	Number of CPU cores to use (1-24 available) (default 24)
  -edge
    	Add edge details to the pattern
  -h int
    	Set the image height (default 1500)
  -i string
    	Input directory containing images for image-based camouflage (default "input")
  -j string
    	Process a JSON file containing a list of color palettes
  -k int
    	Number of main colors for image-based camouflage (default 4)
  -noise
    	Add noise to the pattern
  -o string
    	The output directory for generated images (default "output")
  -r string
    	Color ratios: 'random' for random ratios, integers like '2,1,3' (cycles if fewer than colors) (default: equal)
  -t string
    	Set the pattern type (pat1, pat2, pat3, pat4, pat5, all, or image) (default "pat1")
  -w int
    	Set the image width (default 1500)
```

## JSON Input Format

When using the `-j` flag to process multiple patterns, you need to provide a JSON file containing color palettes. An example `colors.json` file is included in the repository. The format is as follows:

```json
[
  {
    "name": "woodland_sentinel",
    "colors": [
      "#5e8553",
      "#5c4f42",
      "#333330",
      "#c1bc94"
    ]
  },
  {
    "name": "mountain_mist",
    "colors": [
      "#9bb0c1",
      "#c4cecc",
      "#62779d",
      "#414458"
    ]
  }
]
```

