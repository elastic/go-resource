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
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	defaultFileProviderName   = "file"
	defaultFileProviderPrefix = "/"
)

// FileProvider is a provider of files. It can be configured with the prefix
// path where files should be managed.
type FileProvider struct {
	Prefix string
}

// File is a resource that manages a file.
type File struct {
	// Provider is the name of the provider to use, defaults to "file".
	Provider string
	// Path is the path of the file.
	Path string
	// Absent is set to true to indicate that the file should not exist. If it
	// exists, the file is removed.
	Absent bool
	// Mode is the file mode and permissions of the file. If not set, defaults to 0644
	// for files and 0755 for directories.
	Mode *fs.FileMode
	// Directory is set to true to indicate that the file is a directory.
	Directory bool
	// CreateParent is set to true if parent path should be created too.
	CreateParent bool
	// Force forces destructive operations, such as removing a file to replace it
	// with a directory, or the other way around. These operations will fail if
	// force is not set.
	Force bool
	// Content is the content for the file.
	// TODO: Support directory contents.
	Content FileContent
	// KeepExistingContent keeps content of file if it exists.
	KeepExistingContent bool
	// MD5 is the expected md5 sum of the content of the file. If the current content
	// of the file matches this checksum, the file is not updated.
	MD5 string
}

func (f *File) String() string {
	return fmt.Sprintf("[File:%s:%s]", f.Provider, f.Path)
}

func (f *File) provider(ctx Context) *FileProvider {
	name := f.Provider
	if name == "" {
		name = defaultFileProviderName
	}
	var provider *FileProvider
	ok := ctx.Provider(name, &provider)
	if !ok {
		return &FileProvider{Prefix: "/"}
	}
	return provider
}

func (f *File) mode() fs.FileMode {
	switch {
	case f.Mode != nil:
		return *f.Mode
	case f.Directory:
		return 0755
	default:
		return 0644
	}
}

func (f *File) Get(ctx Context) (current ResourceState, err error) {
	provider := f.provider(ctx)
	path := filepath.Join(provider.Prefix, f.Path)
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &FileState{expected: !f.Absent}, nil
	} else if err != nil {
		return nil, err
	}
	return &FileState{
		info:     info,
		expected: !f.Absent,
		context:  ctx,
		content: func() (io.ReadCloser, error) {
			return os.Open(path)
		},
	}, nil
}

func (f *File) Create(ctx Context) error {
	return f.createWithContent(ctx, f.Content)
}

func (f *File) createWithContent(ctx Context, content FileContent) error {
	provider := f.provider(ctx)
	path := filepath.Join(provider.Prefix, f.Path)

	if f.CreateParent {
		err := os.MkdirAll(filepath.Dir(path), f.mode()|0111)
		if err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
	}

	if f.Directory {
		return os.Mkdir(path, f.mode())
	}

	if content != nil {
		err := safeWriteContent(ctx, path, content, f.MD5)
		if err != nil {
			return err
		}
	} else if f.Content == nil {
		created, err := os.Create(path)
		if err != nil {
			return err
		}
		defer created.Close()
	}

	if err := os.Chmod(path, f.mode()); err != nil {
		return fmt.Errorf("failed to set mode: %w", err)
	}

	return nil
}

// safeWriteContent writes the content to a tmp file before overwriting the original file.
// If md5sum is not empty, it checks that the md5 is correct before writing the final file.
func safeWriteContent(ctx Context, path string, content FileContent, md5Sum string) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	checksum := md5.New()
	w := io.MultiWriter(tmpFile, checksum)
	err = content(ctx, w)
	tmpFile.Close()
	if err != nil {
		return err
	}

	if md5Sum != "" && md5Sum != string(checksum.Sum(nil)) {
		return errors.New("md5 checksum of content differs")
	}

	os.Remove(path)
	return os.Rename(tmpFile.Name(), path)
}

func (f *File) Update(ctx Context) error {
	provider := f.provider(ctx)
	path := filepath.Join(provider.Prefix, f.Path)
	if f.Absent {
		return os.Remove(path)
	}
	if f.Force {
		info, err := os.Stat(path)
		if err == nil && info != nil && f.Directory != info.IsDir() {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
	}
	content := f.Content
	if f.KeepExistingContent {
		content = nil
	}
	return f.createWithContent(ctx, content)
}

type FileState struct {
	info     fs.FileInfo
	expected bool
	context  Context
	content  func() (io.ReadCloser, error)
}

func (f *FileState) Found() bool {
	return f.info != nil || !f.expected
}

func (f *FileState) NeedsUpdate(resource Resource) (bool, error) {
	file := resource.(*File)
	if file.Absent && f.info != nil {
		return true, nil
	}
	if f.info != nil && file.Directory != f.info.IsDir() {
		return true, nil
	}
	if f.info != nil && file.mode().Perm() != f.info.Mode().Perm() {
		return true, nil
	}
	if file.Content != nil && !file.KeepExistingContent {
		current, err := f.content()
		if err != nil {
			return true, err
		}
		defer current.Close()

		currentCheckSum := md5.New()
		io.Copy(currentCheckSum, current)
		if file.MD5 != "" && file.MD5 == string(currentCheckSum.Sum(nil)) {
			return false, nil
		}

		expectedCheckSum := md5.New()
		file.Content(f.context, expectedCheckSum)
		if !bytes.Equal(currentCheckSum.Sum(nil), expectedCheckSum.Sum(nil)) {
			return true, nil
		}
	}
	return false, nil
}

// FileMode is a helper function to create a *fs.FileMode inline.
func FileMode(mode fs.FileMode) *fs.FileMode {
	return &mode
}
