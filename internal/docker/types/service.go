package types

import (
	"os"
)

type SharedEvn struct {
	PGPASSWORD string
	JWT_SECRET string
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
	WaitFor() error
}
