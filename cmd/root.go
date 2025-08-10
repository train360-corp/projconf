package cmd

import (
	"github.com/train360-corp/projconf/internal/app"
	"os"
)

func Run() error {
	return app.Get().Run(os.Args)
}
