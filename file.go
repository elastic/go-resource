package resource

import (
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

func (f *File) provider(applyCtx Context) *FileProvider {
	var provider *FileProvider
	ok := applyCtx.Provider(f.Provider, &provider)
	if !ok {
		return &FileProvider{Prefix: "."}
	}
	return provider
}

func (f *File) Get(applyCtx Context) (current Resource, found bool) {
	provider := f.provider(applyCtx)
	_, err := os.Stat(filepath.Join(provider.Prefix, f.Path))
	if err != nil {
		return f, false
	}
	return f, true
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
