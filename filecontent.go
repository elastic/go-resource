package resource

import (
	"fmt"
	"io"
)

type FileContent func(w io.Writer) error

func FileContentLiteral(content string) FileContent {
	return func(w io.Writer) error {
		_, err := fmt.Fprintf(w, content)
		return err
	}
}
