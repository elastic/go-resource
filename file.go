package resource

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type FileProvider struct {
	Prefix string
}

type File struct {
	Provider string
	Path     string
	Content  FileContent
}

func (f *File) String() string {
	return fmt.Sprintf("[File:%s:%s]", f.Provider, f.Path)
}

func (f *File) provider(applyCtx Context) *FileProvider {
	var provider *FileProvider
	ok := applyCtx.Provider(f.Provider, &provider)
	if !ok {
		return &FileProvider{Prefix: "."}
	}
	return provider
}

func (f *File) Get(applyCtx Context) (current ResourceState, found bool) {
	provider := f.provider(applyCtx)
	path := filepath.Join(provider.Prefix, f.Path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return &FileState{
		info: info,
		content: func() (io.ReadCloser, error) {
			return os.Open(path)
		},
	}, true
}

func (f *File) Create(applyCtx Context) error {
	provider := f.provider(applyCtx)
	created, err := os.Create(filepath.Join(provider.Prefix, f.Path))
	if err != nil {
		return err
	}
	defer created.Close()
	if f.Content != nil {
		err := f.Content(created)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *File) Update(applyCtx Context) error {
	return f.Create(applyCtx)
}

type FileState struct {
	info    fs.FileInfo
	content func() (io.ReadCloser, error)
}

func (f *FileState) NeedsUpdate(resource Resource) bool {
	file := resource.(*File)
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
		file.Content(expectedCheckSum)

		if !bytes.Equal(currentCheckSum.Sum(nil), expectedCheckSum.Sum(nil)) {
			return true
		}
	}
	return false
}
