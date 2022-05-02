package resource

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilePresent(t *testing.T) {
	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
	}
	resources := Resources{&resource}

	_, found := resource.Get(manager)
	assert.False(t, found)

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(provider.Prefix, resource.Path))
	assert.NoError(t, err)
}

func TestFileContent(t *testing.T) {
	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	content := "somecontent"
	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Content:  FileContentLiteral(content),
	}
	resources := Resources{&resource}

	_, found := resource.Get(manager)
	assert.False(t, found)

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	require.NoError(t, err)
	assert.Equal(t, content, string(d))
}

func TestFileContentFromSourceFile(t *testing.T) {
	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	source := NewSourceFS(os.DirFS("testdata/templates"))
	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Content:  source.File("sample-file.txt"),
	}
	resources := Resources{&resource}

	_, found := resource.Get(manager)
	assert.False(t, found)

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, "This is a source file.\n", string(d))
	}
}

func TestFileContentFromSourceTemplate(t *testing.T) {
	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	facter := StaticFacter{"sample": "samplefact"}
	manager.AddFacter(facter)

	source := NewSourceFS(os.DirFS("testdata/templates"))
	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Content:  source.Template(manager, "sample-file.txt.tmpl"),
	}
	resources := Resources{&resource}

	_, found := resource.Get(manager)
	assert.False(t, found)

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, "This is a template with a fact: samplefact\n", string(d))
	}
}
