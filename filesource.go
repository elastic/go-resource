package resource

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"text/template"
)

// SourceFS is an abstracted file system that can be used to obtail file contents.
type SourceFS struct {
	fs.FS

	templateFuncs template.FuncMap
}

// NewSourceFS returns a new SourceFS with the root file system.
func NewSourceFS(root fs.FS) *SourceFS {
	return &SourceFS{
		FS: root,
	}
}

// WithTemplateFuncs sets and returns a set of functions that can be used by
// templates in this source file system.
func (s *SourceFS) WithTemplateFuncs(fmap template.FuncMap) *SourceFS {
	s.templateFuncs = fmap
	return s
}

// File returns the file content for a given path in the source file system.
func (s *SourceFS) File(path string) FileContent {
	return func(_ Context, w io.Writer) error {
		f, err := s.FS.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	}

	return nil
}

// Template returns the file content for a given path in the source file system.
// If the file contains a template, this template is executed.
// The template can use the `fact(string) string`  function, as well as other functions
// defined with `WithTemplateFuncs`.
func (s *SourceFS) Template(path string) FileContent {
	return func(applyContext Context, w io.Writer) error {
		fmap := template.FuncMap{
			"fact": func(name string) (string, error) {
				v, found := applyContext.Fact(name)
				if !found {
					return "", fmt.Errorf("fact %q not found", name)
				}
				return v, nil
			},
		}

		t, err := template.New(filepath.Base(path)).Funcs(s.templateFuncs).Funcs(fmap).ParseFS(s.FS, path)
		if err != nil {
			return err
		}
		return t.Execute(w, nil)
	}
}

// DefaultHTTPSource is a SourceHTTP that uses the default HTTP client.
var DefaultHTTPSource = &HTTPSource{Client: http.DefaultClient}

// HTTPSource is a file source that can be used to obtain contents from http resources.
type HTTPSource struct {
	// Client is the client used to make HTTP requests. If no client is configured,
	// the default one is used.
	Client *http.Client
}

// Get obtains the content with an http request to the given location.
func (s *HTTPSource) Get(location string) FileContent {
	return func(_ Context, w io.Writer) error {
		client := s.Client
		if client == nil {
			client = http.DefaultClient
		}
		resp, err := client.Get(location)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, err = io.Copy(w, resp.Body)
		return err
	}
}
