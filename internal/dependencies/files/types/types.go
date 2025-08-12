package types

import "os"

type Writeable struct {
	Path string
	Data []byte
	Perm os.FileMode
}

type EmbeddedFile interface {
	GetDisplay() string
	GetWriteables() []Writeable
}
