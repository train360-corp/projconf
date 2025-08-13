package types

import (
	"context"
	"os"
)

type SharedEvn struct {
	PGPASSWORD  string
	JWT_SECRET  string
	ANON_KEY    string
	SERVICE_KEY string
}

type Writeable struct {
	LocalPath     string
	Data          []byte
	Perm          os.FileMode
	ContainerPath string
}

type Service interface {
	GetDisplay() string
	Run(evn *SharedEvn) error
	GetWriteables() []Writeable
	WaitFor(ctx context.Context) error
}
