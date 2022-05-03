package resource

import (
	"fmt"
	"io"
)

type FileContent func(Context, io.Writer) error

func FileContentLiteral(content string) FileContent {
	return func(_ Context, w io.Writer) error {
		_, err := fmt.Fprintf(w, content)
		return err
	}
}
