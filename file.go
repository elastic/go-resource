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

type FileProvider struct {
	Prefix string
}

type File struct {
	Provider string
	Path     string
	Content  FileContent
	Absent   bool
}

func (f *File) String() string {
	return fmt.Sprintf("[File:%s:%s]", f.Provider, f.Path)
}

func (f *File) provider(applyCtx Context) *FileProvider {
	name := f.Provider
	if name == "" {
		name = defaultFileProviderName
	}
	var provider *FileProvider
	ok := applyCtx.Provider(name, &provider)
	if !ok {
		return &FileProvider{Prefix: "/"}
	}
	return provider
}

func (f *File) Get(applyCtx Context) (current ResourceState, err error) {
	provider := f.provider(applyCtx)
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
		context:  applyCtx,
		content: func() (io.ReadCloser, error) {
			return os.Open(path)
		},
	}, nil
}

func (f *File) Create(applyCtx Context) error {
	provider := f.provider(applyCtx)
	created, err := os.Create(filepath.Join(provider.Prefix, f.Path))
	if err != nil {
		return err
	}
	defer created.Close()
	if f.Content != nil {
		err := f.Content(applyCtx, created)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *File) Update(applyCtx Context) error {
	if f.Absent {
		provider := f.provider(applyCtx)
		return os.Remove(filepath.Join(provider.Prefix, f.Path))
	}
	return f.Create(applyCtx)
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

func (f *FileState) NeedsUpdate(resource Resource) bool {
	file := resource.(*File)
	if file.Absent && f.info != nil {
		return true
	}
	if file.Content != nil {
		current, err := f.content()
		if err != nil {
			// TODO: improve error handling here.
			return true
		}
		defer current.Close()

		currentCheckSum := md5.New()
		io.Copy(currentCheckSum, current)

		expectedCheckSum := md5.New()
		file.Content(f.context, expectedCheckSum)

		if !bytes.Equal(currentCheckSum.Sum(nil), expectedCheckSum.Sum(nil)) {
			return true
		}
	}
	return false
}
