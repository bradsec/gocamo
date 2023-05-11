package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type camoColors struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

func showBanner() {
	bannerArt := `
 ######   ######   ######  #####  ###    ###  ###### 
##       ##    ## ##      ##   ## ####  #### ##    ##
##   ### ##    ## ##      ####### ## #### ## ##    ##
##    ## ##    ## ##      ##   ## ##  ##  ## ##    ##
 ######   ######   ###### ##   ## ##      ##  ###### 
`
	fmt.Println(bannerArt)
}

func showUsage() {
	usageText := `Usage:

gocamo [OPTIONS]

Available options:

-r:  Generate a single pattern with random colors.
     Example: gocamo -r

-c:  Generate a single pattern using a comma-separated list of hex colors.
     Example: gocamo -c "#ffffff,#012169,#e4002b"

-j:  Batch process a JSON file containing a list of color palettes.
     Example: gocamo -j "colors.json"

-w:  Set the image width (default is 3840).
-h:  Set the image height (default is 2160).
     Example: gocamo -j colors.json -w 1920 -h 1080

-o:  The output directory or folder for generated images (default creates an 'output' in current directory).
     Example: gocamo -j colors.json -o thisfolder

-lg: Set the size of the large squares in the pattern (default is 20).
-sm: Set the size of the small squares in the pattern (default is 10).
     Example: gocamo -r -lg 50 -sm 25
`
	fmt.Println(usageText)
}

func hexToRGBA(hex string) (color.RGBA, error) {
	hex = strings.ReplaceAll(hex, "#", "")
	r, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}

func randomHexColor() string {
	r := rand.Intn(256)
	g := rand.Intn(256)
	b := rand.Intn(256)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func randomCamoColors() camoColors {
	colors := []string{}
	for i := 0; i < 4; i++ { // generate 4 colors
		colors = append(colors, randomHexColor())
	}
	return camoColors{
		Name:   "random",
		Colors: colors,
	}
}

func getAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Error getting absolute path to '%s': %v", path, err)
	}
	return absPath, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generateCamo(camo camoColors, ci int, total int, width int, height int, squareSize int, smallSquareSize int, outputPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Generate file name with name of camo and hex codes of colors used in image
	fileName := fmt.Sprintf("gocamo_%03d_", ci)
	for hi, hex := range camo.Colors {
		hex = strings.ReplaceAll(hex, "#", "")
		if hi != len(camo.Colors)-1 {
			fileName += fmt.Sprintf("%s_", hex)
		} else {
			fileName += fmt.Sprintf("%s", hex)
		}
	}
	fileName += fmt.Sprintf("_w%vxh%v", width, height)
	fileName += ".png"

	// Convert hex colors to color.RGBA format
	colors := make([]color.RGBA, len(camo.Colors))
	for i, hex := range camo.Colors {
		rgba, err := hexToRGBA(hex)
		if err != nil {
			fmt.Printf("Error converting hex to RGBA: %v", err)
			return
		}
		colors[i] = rgba
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create an index to keep track of the last color used
	lastColorIndex := -1

	for x := 0; x < width; x += squareSize {
		for y := 0; y < height; y += squareSize {
			// Determine bleed factor for color
			bleedFactor := rand.Intn(squareSize / 2)

			// Determine size of color group
			colorGroupSize := squareSize + bleedFactor

			// Determine color for group
			// 45% chance to use the last color again
			var color color.RGBA
			if lastColorIndex != -1 && rand.Float32() < 0.45 {
				color = colors[lastColorIndex]
			} else {
				lastColorIndex = rand.Intn(len(colors))
				color = colors[lastColorIndex]
			}

			// Fill in color group with selected color
			for i := x; i < x+colorGroupSize; i++ {
				for j := y; j < y+colorGroupSize; j++ {
					if i < width && j < height {
						img.Set(i, j, color)
					}
				}
			}
		}
	}

	// Generate second layer pattern with smaller squares and less coverage
	for x := 0; x < width; x += smallSquareSize {
		for y := 0; y < height; y += smallSquareSize {
			if rand.Float32() < 0.5 { // 50% chance to draw a small square
				// Determine bleed factor for color
				bleedFactor := rand.Intn(smallSquareSize / 4) // limit bleedFactor to a smaller value

				// Determine size of color group
				colorGroupSize := smallSquareSize + bleedFactor

				// Determine color for group
				color := colors[rand.Intn(len(colors))]

				// Fill in color group with selected color
				for i := x; i < x+colorGroupSize; i++ {
					for j := y; j < y+colorGroupSize; j++ {
						if i < width && j < height {
							img.Set(i, j, color)
						}
					}
				}
			}
		}
	}

	// Save image
	filePath := filepath.Join(outputPath, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	png.Encode(f, img)
	fmt.Printf("\r%s", fileName)
}

func main() {

	// Terminal display helper variables
	// lineUp := "\033[1A"
	lineClear := "\x1b[2K"

	// Show banner art
	showBanner()

	// Define command line arguments
	widthPtr := flag.Int("w", 3840, "Set the image width.")
	heightPtr := flag.Int("h", 2160, "Set the image height.")
	lgSquareSizePtr := flag.Int("lg", 20, "Set the size of the large squares in the pattern.")
	smSquareSizePtr := flag.Int("sm", 10, "Set the size of the small squares in the pattern.")
	jsonFilePtr := flag.String("j", "colors.json", "Batch process a JSON file containing a list of color palettes.")
	colorsPtr := flag.String("c", "", "Generate a single pattern using a comma-separated list of hex colors.")
	outputPtr := flag.String("o", "output", "The output directory or folder for generated images.")
	randomPtr := flag.Bool("r", false, "Generate a single pattern with random colors.")

	// Parse command line arguments
	flag.Parse()

	if flag.NFlag() == 0 {
		showUsage()
		return
	}

	// Generate pattern
	fmt.Printf("[GENERATING] Digital Camouflage Patterns in %vx%v...\n", *widthPtr, *heightPtr)

	// Get absolute path to output directory
	outputAbsPath, err := getAbsPath(*outputPtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check if output directory exists
	if _, err := os.Stat(outputAbsPath); os.IsNotExist(err) {
		// Output directory does not exist, prompt user to create it
		fmt.Printf("The output directory '%s' does not exist. Do you want to create it? (y/n): ", outputAbsPath)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			fmt.Printf("Error reading user input: %v\n", scanner.Err())
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" || strings.ToLower(input) != "y" {
			fmt.Println("Exiting program.")
			return
		}
		// Create output directory
		if err := os.Mkdir(outputAbsPath, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}
	}

	// Start timing
	startTime := time.Now()

	// If colors are specified, do one-off generation
	if *colorsPtr != "" {
		colors := strings.Split(*colorsPtr, ",")
		camo := camoColors{
			Colors: colors,
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go generateCamo(camo, 0, 1, *widthPtr, *heightPtr, *lgSquareSizePtr, *smSquareSizePtr, *outputPtr, &wg)
		wg.Wait()
		// End timing and print duration
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		seconds := duration.Seconds()
		fmt.Printf("%v\r[GENERATION] Completed in %.2f seconds.\n[GENERATION] File saved in %v\n", lineClear, seconds, outputAbsPath)
		return
	}

	if *randomPtr {
		camo := randomCamoColors()
		var wg sync.WaitGroup
		wg.Add(1)
		go generateCamo(camo, 0, 1, *widthPtr, *heightPtr, *lgSquareSizePtr, *smSquareSizePtr, *outputPtr, &wg)
		wg.Wait()
		// End timing and print duration
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		seconds := duration.Seconds()
		fmt.Printf("%v\r[GENERATION] Completed in %.2f seconds.\n[GENERATION] File saved in %v\n", lineClear, seconds, outputAbsPath)
		return
	}

	// Get absolute path to JSON file
	jsonAbsPath, err := getAbsPath(*jsonFilePtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check if JSON file exists
	if _, err := os.Stat(jsonAbsPath); os.IsNotExist(err) {
		fmt.Printf("The JSON file '%s' does not exist.\n", jsonAbsPath)
		return
	}

	// Read JSON file
	file, err := os.Open(*jsonFilePtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Parse JSON file into slice of camoColors
	var camoList []camoColors
	err = json.NewDecoder(file).Decode(&camoList)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Display total palettes in JSON file
	fmt.Printf("[GENERATING] Processing %v color palettes from %v...\n", len(camoList), *jsonFilePtr)

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Iterate over camo colors
	for ci, camo := range camoList {
		// Add one to the WaitGroup counter
		wg.Add(1)

		// Start a new goroutine to generate the camouflage pattern
		go generateCamo(camo, ci, len(camoList), *widthPtr, *heightPtr, *lgSquareSizePtr, *smSquareSizePtr, *outputPtr, &wg)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// End timing and print duration
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	seconds := duration.Seconds()
	fmt.Printf("%v\r[GENERATION] Completed in %.2f seconds.\n[GENERATION] Files saved in %v\n", lineClear, seconds, outputAbsPath)
}
