package swaggerFiles

import (
	"embed"
	"io/fs"
	"log"
)

// swagger-ui-dist v5.33.1, see dist/LICENSE.
//
//go:embed dist
var dist embed.FS

func init() {
	err := fs.WalkDir(dist, "dist", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := dist.ReadFile(name)
		if err != nil {
			return err
		}
		return WriteFile(name[len("dist"):], data, 0644)
	})
	if err != nil {
		log.Fatal(err)
	}
}
