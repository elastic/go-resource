package resource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	prefix := t.TempDir()
	fileName := "somefile.txt"

	cmd := Main{
		Facters: []Facter{
			&EnvFacter{},
		},
		Providers: map[string]Provider{
			"file": &FileProvider{
				Prefix: prefix,
			},
		},
		Resources: []Resource{
			&File{
				Path: fileName,
			},
		},
	}

	err := cmd.Run()
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(prefix, fileName))
	assert.NoError(t, err)
}
