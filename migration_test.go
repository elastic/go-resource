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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration(t *testing.T) {
	provider := FileProvider{
		Prefix: t.TempDir(),
	}
	f, err := os.Create(filepath.Join(provider.Prefix, "old-file.txt"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	manager := NewManager()
	manager.RegisterProvider(defaultFileProviderName, &provider)

	migrator := NewMigrator(&dummyVersion{1})
	migrator.AddMigration(1, func(m *Manager) (ApplyResults, error) {
		t.Fatal("this migration should not be called")
		return nil, fmt.Errorf("failed")
	})
	migrator.AddMigration(2, func(m *Manager) (ApplyResults, error) {
		return m.Apply(Resources{
			&File{
				Path:   "old-file.txt",
				Absent: true,
			},
		})
	})
	manager.withMigrator(migrator)

	results, err := manager.Apply(Resources{
		&File{
			Path: "new-file.txt",
		},
	})
	assert.NoError(t, err)
	if assert.Len(t, results, 2) {
		assert.Equal(t, ActionUpdate, results[0].action)
		assert.Equal(t, ActionCreate, results[1].action)
	}

	results, err = manager.Apply(Resources{
		&File{
			Path: "new-file.txt",
		},
	})
	assert.Empty(t, results)
}

type dummyVersion struct {
	version uint
}

func (v *dummyVersion) Current() uint { return v.version }
func (v *dummyVersion) Set(version uint) error {
	v.version = version
	return nil
}
