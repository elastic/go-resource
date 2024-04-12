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
	"fmt"
	"io"
)

// FileContent defines the content of a file. It recives an apply context
// to obtain information from the execution, and a writer where to write
// the content.
type FileContent func(context.Context, io.Writer) error

// FileContentLiteral returns a literal file content.
func FileContentLiteral(content string) FileContent {
	return func(_ context.Context, w io.Writer) error {
		_, err := fmt.Fprint(w, content)
		return err
	}
}
