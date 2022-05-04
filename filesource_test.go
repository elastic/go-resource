package resource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	state, err := resource.Get(manager)
	require.NoError(t, err)
	assert.False(t, state.Found())

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

	funcs := template.FuncMap{
		"sayHello": func() string { return "Hello!" },
	}
	source := NewSourceFS(os.DirFS("testdata/templates")).WithTemplateFuncs(funcs)
	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Content:  source.Template("sample-file.txt.tmpl"),
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager)
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, "Hello! This is a template with a fact: samplefact\n", string(d))
	}
}

func TestFileContentFromSourceURL(t *testing.T) {
	expectedContent := "Some content from the Internet!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedContent)
	}))

	providerName := "test-files"
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	manager := NewManager()
	manager.RegisterProvider(providerName, &provider)

	resource := File{
		Provider: providerName,
		Path:     "/sample-file.txt",
		Content:  DefaultHTTPSource.Get(server.URL),
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager)
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := ioutil.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, expectedContent, string(d))
	}
}
