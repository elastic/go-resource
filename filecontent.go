package resource

import (
	"fmt"
	"io"
)

// FileContent defines the content of a file. It recives an apply context
// to obtain information from the execution, and a writer where to write
// the content.
type FileContent func(Context, io.Writer) error

// FileContentLiteral returns a literal file content.
func FileContentLiteral(content string) FileContent {
	return func(_ Context, w io.Writer) error {
		_, err := fmt.Fprintf(w, content)
		return err
	}
}
