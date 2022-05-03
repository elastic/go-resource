package resource

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"text/template"
)

type SourceFS struct {
	root          fs.FS
	templateFuncs template.FuncMap
}

func NewSourceFS(root fs.FS) *SourceFS {
	return &SourceFS{
		root: root,
	}
}

func (s *SourceFS) WithTemplateFuncs(fmap template.FuncMap) *SourceFS {
	s.templateFuncs = fmap
	return s
}

func (s *SourceFS) File(path string) FileContent {
	return func(_ Context, w io.Writer) error {
		f, err := s.root.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	}

	return nil
}

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

		t, err := template.New(filepath.Base(path)).Funcs(s.templateFuncs).Funcs(fmap).ParseFS(s.root, path)
		if err != nil {
			return err
		}
		return t.Execute(w, nil)
	}
}
