package files

import (
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/dependencies/files/supabase"
	"github.com/train360-corp/projconf/internal/dependencies/files/types"
	"github.com/train360-corp/projconf/internal/fs"
)

func WriteTempFiles() error {

	embeddables := []types.EmbeddedFile{
		supabase.DatabaseEmbed{},
	}

	for _, embeddable := range embeddables {
		for _, writeable := range embeddable.GetWriteables() {
			//log.Printf("wrote temp file: %s\n", writeable.Path)
			if err := fs.WriteTempFile(writeable.Path, writeable.Data, writeable.Perm); err != nil {
				return errors.New(fmt.Sprintf("write temp file failed (path=%s): %s", writeable.Path, err))
			}
		}
	}

	return nil
}
