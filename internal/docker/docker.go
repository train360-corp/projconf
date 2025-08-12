package docker

import (
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/docker/utils"
	"log"
)

func RunDockerServices(env types.SharedEvn) error {

	services := []types.Service{
		database.Service{},
	}

	for _, service := range services {
		log.Printf("starting %s...\n", service.GetDisplay())
		if err := utils.WriteTempFiles(service.GetWriteables()); err != nil {
			return errors.New(fmt.Sprintf("failed to write temp files for service \"%s\": %s", service.GetDisplay(), err.Error()))
		}
		err := service.Run(&env)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to run service \"%s\": %s", service.GetDisplay(), err.Error()))
		}
	}

	return nil
}
