package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bradsec/gocamo/internal/generator"
	"github.com/bradsec/gocamo/pkg/config"
)

type Job struct {
	Camo       config.CamoColors
	ImagePath  string
	Index      int
	Config     *config.Config
	OutputPath string
}

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
