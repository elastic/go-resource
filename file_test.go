package resource

import (
	"errors"
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

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	assert.Equal(t, ActionCreate, result[0].action)

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

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	assert.Equal(t, ActionCreate, result[0].action)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	require.NoError(t, err)
	assert.Equal(t, content, string(d))
}

func TestFileContentUpdate(t *testing.T) {
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

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.False(t, state.Found())

	err = ioutil.WriteFile(filepath.Join(provider.Prefix, resource.Path), []byte("old content"), 0644)
	require.NoError(t, err)

	state, err = resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.True(t, state.Found())

	// On first apply, it should update the content to the expected one.
	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	if assert.NotEmpty(t, result, "expecting update") {
		assert.Equal(t, ActionUpdate, result[0].action)
	}

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	require.NoError(t, err)
	assert.Equal(t, content, string(d))

	// On second apply, it should do nothing.
	result, err = manager.Apply(resources)
	t.Log(result)
	require.Empty(t, result)
}

func TestFileDefaultProvider(t *testing.T) {
	manager := NewManager()

	resource := File{
		Path: filepath.Join(t.TempDir(), "sample-file.txt"),
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	assert.Equal(t, ActionCreate, result[0].action)

	_, err = os.Stat(resource.Path)
	assert.NoError(t, err)
}

func TestFileOverrideDefaultProvider(t *testing.T) {
	providerName := defaultFileProviderName
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	resource := File{
		Path: "/sample-file.txt",
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	assert.Equal(t, ActionCreate, result[0].action)

	_, err = os.Stat(filepath.Join(provider.Prefix, resource.Path))
	assert.NoError(t, err)
}

func TestFileAbsent(t *testing.T) {
	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Absent:   true,
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.True(t, state.Found())

	f, err := os.Create(filepath.Join(provider.Prefix, resource.Path))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	state, err = resource.Get(manager.Context(nil))
	require.NoError(t, err)
	assert.True(t, state.Found())

	// On first apply, it should remove the file.
	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)
	if assert.NotEmpty(t, result, "expecting update") {
		assert.Equal(t, ActionUpdate, result[0].action)
	}

	_, err = os.Stat(filepath.Join(provider.Prefix, resource.Path))
	require.True(t, errors.Is(err, os.ErrNotExist))

	// On second apply, it should do nothing.
	result, err = manager.Apply(resources)
	t.Log(result)
	require.Empty(t, result)
}
