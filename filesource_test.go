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
	"context"
	"crypto/md5"
	"fmt"
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

	state, err := resource.Get(manager.Context(context.Background()))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := os.ReadFile(filepath.Join(provider.Prefix, resource.Path))
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

	state, err := resource.Get(manager.Context(context.Background()))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := os.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, "Hello! This is a template with a fact: samplefact\n", string(d))
	}
}

func TestFileContentFromSourceURL(t *testing.T) {
	expectedContent := "Some content from the Internet!"
	expectedMD5 := md5.Sum([]byte(expectedContent))
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
		MD5:      string(expectedMD5[:]),
	}
	resources := Resources{&resource}

	state, err := resource.Get(manager.Context(context.Background()))
	require.NoError(t, err)
	assert.False(t, state.Found())

	result, err := manager.Apply(resources)
	t.Log(result)
	require.NoError(t, err)

	d, err := os.ReadFile(filepath.Join(provider.Prefix, resource.Path))
	if assert.NoError(t, err) {
		assert.Equal(t, expectedContent, string(d))
	}
}
