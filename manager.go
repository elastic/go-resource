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
	"reflect"
)

// Provider is the interface implemented by providers.
type Provider interface {
}

// Facter is the interface implemented by facters.
// Facters provide, facts, with information about the execution context, they
// can be queried through the manager.
type Facter interface {
	// Fact returns the value of a fact for a given name and true if it is found.
	// It not found, it returns an empty string and false.
	Fact(name string) (value string, found bool)
}

// StaticFacter is a facter implemented as map.
type StaticFacter map[string]string

// Fact returns the value of a fact for a given name and true if it is found.
// It not found, it returns an empty string and false.
func (f StaticFacter) Fact(name string) (value string, found bool) {
	if f == nil {
		return "", false
	}
	value, found = f[name]
	return
}

// Resource implements management for a resource.
type Resource interface {
	// Get gets the current state of a resource. An error is returned if the state couldn't
	// be determined. An error here interrupts execution.
	Get(context.Context) (current ResourceState, err error)

	// Create implements the creation of the resource. It can return an error, that is reported
	// as part of the execution result.
	Create(context.Context) error

	// Update implements the upodate of an existing resource. Ir can return an error, that
	// is reported as part of the execution result.
	Update(context.Context) error
}

// ResourceState is the state of a resource.
type ResourceState interface {
	// Found returns true if the resource exists.
	Found() bool

	// NeedsUpdate returns true if the resource needs update when compared with the given
	// resource definition.
	NeedsUpdate(definition Resource) (bool, error)
}

// Resources is a collection of resources.
type Resources []Resource

// Actions reported on results when applying resources.
const (
	// ActionUnknown is used to indicate a failure happening before determining the required action.
	ActionUnknown = "unkwnown"

	// ActionCreate refers to an action that creates a resource.
	ActionCreate = "create"

	// ActionUpdate refers to an action that affects an existing resource.
	ActionUpdate = "update"
)

// ApplyResult is the result of applying a resource.
type ApplyResult struct {
	action   string
	resource Resource
	err      error
}

// Err returns an error if the application of a resource failed.
func (r ApplyResult) Err() error {
	return r.err
}

// String returns the string representation of the result of applying a resource.
func (r ApplyResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("{%s: %s, failed: %v}", r.action, r.resource, r.err)
	} else {
		return fmt.Sprintf("{%s: %s}", r.action, r.resource)
	}
}

// ApplyResults is the colection of results when applying a collection of resources.
type ApplyResults []ApplyResult

// Runtime is the context of execution when applying resources.
type Runtime interface {
	// Provider obtains a provider from the runtime, and sets it in the target.
	// The target must be a pointer to a provider type.
	// It returns false, and doesn't set the target if no provider is found with
	// the given name and target type.
	Provider(name string, target any) (found bool)

	// Fact returns the value of a fact for a given name and true if it is found.
	// It not found, it returns an empty string and false.
	Fact(name string) (value string, found bool)
}

// Manager manages application of resources, it contains references to providers and
// facters.
type Manager struct {
	providers map[string]Provider
	facters   []Facter

	// TBD: pending to confirm migrating API
	migrator *Migrator
}

// NewManager instantiates a new empty manager.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// contextKey is a custom type to avoid collisions in key values.
type contextKey int

const contextRuntimeKey contextKey = iota

// ContextWithRuntime returns a resource context that wraps the given context and the manager.
func (m *Manager) ContextWithRuntime(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextRuntimeKey, m)
}

// RuntimeFromContext obtains a runtime from a context. The context must contain a runtime,
// otherwise this call will panic.
func RuntimeFromContext(ctx context.Context) Runtime {
	v := ctx.Value(contextRuntimeKey)
	runtime, ok := v.(Runtime)
	if !ok {
		panic("context without resources runtime")
	}
	return runtime
}

// Register provider registers a provider in the Manager.
func (m *Manager) RegisterProvider(name string, provider Provider) {
	m.providers[name] = provider
}

// withMigrator sets a migrator in the manager.
// TBD: not exposed, pending to confirm migrating API
func (m *Manager) withMigrator(migrator *Migrator) {
	m.migrator = migrator
}

// Provider obtains a provider from the context, and sets it in the target.
// The target must be a pointer to a provider type.
// It returns false, and doesn't set the target if no provider is found with
// the given name and target type.
func (m *Manager) Provider(name string, target any) bool {
	if target == nil {
		panic("target provider shound not be nil")
	}
	p, found := m.providers[name]
	if !found {
		return false
	}
	val := reflect.ValueOf(target)
	if !reflect.TypeOf(p).AssignableTo(val.Elem().Type()) {
		return false
	}
	val.Elem().Set(reflect.ValueOf(p))
	return true
}

// Apply applies a collection of resources. Depending on their current state,
// resources are created or updated.
func (m *Manager) Apply(resources Resources) (ApplyResults, error) {
	return m.ApplyCtx(context.Background(), resources)
}

// ApplyCtx applies a collection of resources with a context that is passed to resource
// operations.
// Depending on their current state, resources are created or updated.
func (m *Manager) ApplyCtx(ctx context.Context, resources Resources) (ApplyResults, error) {
	results, err := m.applyMigrations()
	if err != nil {
		return results, fmt.Errorf("migrator failed: %w", err)
	}

	resourceResults, err := m.applyResources(ctx, resources)
	results = append(results, resourceResults...)
	return results, err
}

// applyMigrations applies the configured migrations.
func (m *Manager) applyMigrations() (ApplyResults, error) {
	if m.migrator == nil {
		return nil, nil
	}

	// Avoid infinite loops.
	managerWithoutMigrator := &Manager{
		providers: m.providers,
		facters:   m.facters,
	}
	return m.migrator.RunMigrations(managerWithoutMigrator)
}

// applyResources applies a collection of resources. Depending on their current
// state, resources are created or updated.
func (m *Manager) applyResources(ctx context.Context, resources Resources) (ApplyResults, error) {
	applyCtx := m.ContextWithRuntime(ctx)
	var results ApplyResults
	var errors []error
	for _, resource := range resources {
		if err := applyCtx.Err(); err != nil {
			errors = append(errors, fmt.Errorf("apply interrupted: %w", err))
			break
		}

		result := m.applyResource(applyCtx, resource)
		if result == nil {
			continue
		}
		if result.err != nil {
			errors = append(errors, result.err)
		}
		results = append(results, *result)
	}
	return results, newApplyError(errors)
}

// applyResource is a helper function that applies a single resource.
func (m *Manager) applyResource(ctx context.Context, resource Resource) *ApplyResult {
	current, err := resource.Get(ctx)
	if err != nil {
		return &ApplyResult{
			action:   ActionUnknown,
			resource: resource,
			err:      err,
		}
	}

	if !current.Found() {
		err := resource.Create(ctx)
		return &ApplyResult{
			action:   ActionCreate,
			resource: resource,
			err:      err,
		}
	}

	needsUpdate, err := current.NeedsUpdate(resource)
	if err != nil {
		return &ApplyResult{
			action:   ActionUnknown,
			resource: resource,
			err:      err,
		}
	}
	if needsUpdate {
		err := resource.Update(ctx)
		return &ApplyResult{
			action:   ActionUpdate,
			resource: resource,
			err:      err,
		}
	}

	// No action applied to this resource.
	return nil
}

// AddFacter adds a facter to the manager. Facters added later have precedence.
func (m *Manager) AddFacter(facter Facter) {
	m.facters = append([]Facter{facter}, m.facters...)
}

// Fact returns the value of a fact for a given name and true if it is found.
// It not found, it returns an empty string and false.
// If a fact is available in multiple facters, the value in the last added facter
// is returned.
func (m *Manager) Fact(name string) (string, bool) {
	for _, facter := range m.facters {
		v, found := facter.Fact(name)
		if found {
			return v, true
		}
	}
	return "", false
}

// applyError wraps all the errors happened while applying a set of resources.
// Errors can be unwrapped with `Unwrap() []error`.
type applyError struct {
	errors []error
}

// newApplyError returns an error wrapping all the given errors, or nil if
// there were no error.
func newApplyError(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	return &applyError{errors: errors}
}

// Error implements the error interface.
func (e *applyError) Error() string {
	if len(e.errors) == 1 {
		return fmt.Sprintf("there was an apply error: %s", e.errors[0].Error())
	}
	return fmt.Sprintf("there were %d errors", len(e.errors))
}

// Unwrap allows to access wrapped errors.
func (e *applyError) Unwrap() []error {
	return e.errors
}
