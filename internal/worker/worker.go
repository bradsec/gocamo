package worker

import (
	"context"
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
		if j.Config.PatternType == "image" {
			err = generator.GenerateFromImage(ctx, j.Config, j.ImagePath, j.Index, j.OutputPath)
		} else {
			err = generator.GeneratePattern(ctx, j.Config, j.Camo, j.Index, j.OutputPath)
		}
		cancel()
		results <- err
	}
}
