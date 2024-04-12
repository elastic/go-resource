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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManagerProviders(t *testing.T) {
	type TestProvider struct {
		Something string
	}

	provider := &TestProvider{Something: "foo"}

	m := NewManager()
	m.RegisterProvider("test", provider)

	var found *TestProvider
	ok := m.Provider("test", &found)
	if assert.True(t, ok) {
		assert.Equal(t, provider.Something, found.Something)
	}
}

func TestManagerContextCancelled(t *testing.T) {
	t.Run("cancelled before calling to apply", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		m := NewManager()
		resources := Resources{
			&dummyResource{},
		}
		_, err := m.ApplyCtx(cancelledCtx, resources)
		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, context.Canceled))
		}
	})
	t.Run("cancelled on resource creation", func(t *testing.T) {
		m := NewManager()
		resources := Resources{
			&dummyResource{
				absent:      true,
				createError: fmt.Errorf("could not create resource: %w", context.Canceled),
			},
		}
		_, err := m.ApplyCtx(context.Background(), resources)
		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, context.Canceled))
		}
	})
}

func TestApplyError(t *testing.T) {
	t.Run("nil error on empty list", func(t *testing.T) {
		err := newApplyError([]error{})
		assert.NoError(t, err)
	})
	t.Run("propagate context errors", func(t *testing.T) {
		err := newApplyError([]error{context.Canceled})
		assert.True(t, errors.Is(err, context.Canceled))
		assert.Equal(t, "there was an apply error: context canceled", err.Error())
	})
	t.Run("propagate wrapped context errors", func(t *testing.T) {
		err := newApplyError([]error{
			errors.New("some error"),
			fmt.Errorf("interrupted: %w", context.Canceled),
		})
		assert.True(t, errors.Is(err, context.Canceled))
		assert.Equal(t, "there were 2 errors", err.Error())
	})
}

type dummyResource struct {
	absent      bool
	needsUpdate bool
	createError error
}

func (r *dummyResource) Get(context.Context) (ResourceState, error) {
	return &dummyResourceState{
		absent:      r.absent,
		needsUpdate: r.needsUpdate,
	}, nil
}
func (r *dummyResource) Create(context.Context) error { return r.createError }
func (r *dummyResource) Update(context.Context) error { return nil }

type dummyResourceState struct {
	absent      bool
	needsUpdate bool
}

func (s *dummyResourceState) Found() bool { return !s.absent }
func (s *dummyResourceState) NeedsUpdate(definition Resource) (bool, error) {
	return s.needsUpdate, nil
}
