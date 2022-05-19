// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
	return func(ctx Context, w io.Writer) error {
		client := s.Client
		if client == nil {
			client = http.DefaultClient
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, err = io.Copy(w, resp.Body)
		return err
	}
}
