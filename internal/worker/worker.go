package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bradsec/gocamo/internal/generator"
	"github.com/bradsec/gocamo/pkg/config"
)

// Job represents a camouflage pattern generation task that can be processed by a worker.
// It contains all the necessary information to generate either a color palette-based or image-based pattern.
type Job struct {
	Camo       config.CamoColors
	ImagePath  string
	Index      int
	Config     *config.Config
	OutputPath string
}

// Work processes jobs from a channel and executes camouflage pattern generation tasks.
// This function is designed to run as a goroutine in a worker pool pattern. It continuously
// receives jobs from the jobs channel and processes them until the channel is closed.
//
// Each job is executed with a 60-second timeout to prevent hanging operations. The function
// supports two types of pattern generation:
//   - Color palette-based patterns using predefined color schemes
//   - Image-based patterns that extract colors from source images
//
// The pattern type is determined by the Config.PatternType field in the job. All operations
// are performed with proper timeout handling and context cancellation support.
//
// Parameters:
//   - jobs: A receive-only channel providing Job structs to process
//   - results: A send-only channel where job completion results are sent (nil for success, error for failure)
//   - wg: A WaitGroup used to coordinate worker lifecycle and ensure proper shutdown
//
// The function decrements the WaitGroup counter when it exits and sends a result
// (success or error) for each processed job to the results channel.
func Work(jobs <-chan Job, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		var err error

		done := make(chan error, 1)
		go func() {
			if j.Config.PatternType == "image" {
				done <- generator.GenerateFromImage(ctx, j.Config, j.ImagePath, j.Index, j.OutputPath)
			} else {
				done <- generator.GeneratePattern(ctx, j.Config, j.Camo, j.Index, j.OutputPath)
			}
		}()

		select {
		case err = <-done:
		case <-ctx.Done():
			err = fmt.Errorf("operation timed out")
		}

		cancel()
		results <- err
	}
}
