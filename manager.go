package resource

import (
	"errors"
	"reflect"
)

type Provider interface {
}

type Facter interface {
	Fact(name string) (value string, found bool)
}

type StaticFacter map[string]string

func (f StaticFacter) Fact(name string) (value string, found bool) {
	if f == nil {
		return "", false
	}
	value, found = f[name]
	return
}

type Resource interface {
	Get(Context) (current Resource, found bool)
	Create(Context) error
	Update(Context) error
}

type Resources []Resource

const (
	ActionCreate = "create"
	ActionUpdate = "update"
)

type ApplyResult struct {
	action   string
	resource Resource
	err      error
}

func (r *ApplyResult) Err() error {
	return r.err
}

type ApplyResults []ApplyResult

type Context interface {
	Provider(name string, target any) (found bool)
	Fact(name string) (value string, found bool)
}

type Manager struct {
	providers map[string]Provider
	facters   []Facter
}

func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

func (m *Manager) RegisterProvider(name string, provider Provider) {
	m.providers[name] = provider
}

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

func (m *Manager) Apply(resources Resources) (ApplyResults, error) {
	var results ApplyResults
	for _, resource := range resources {
		current, found := resource.Get(m)
		if !found {
			err := resource.Create(m)
			results = append(results, ApplyResult{
				action:   ActionCreate,
				resource: resource,
				err:      err,
			})
			continue
		}

		if !areEqual(current, resource) {
			err := resource.Update(m)
			results = append(results, ApplyResult{
				action:   ActionUpdate,
				resource: resource,
				err:      err,
			})
			continue
		}
	}
	var err error
	for _, result := range results {
		if result.Err() != nil {
			err = errors.New("there where errors")
			break
		}
	}
	return results, err
}

func areEqual(a, b Resource) bool {
	// TODO
	return false
}

func (m *Manager) AddFacter(facter Facter) {
	m.facters = append([]Facter{facter}, m.facters...)
}

func (m *Manager) Fact(name string) (string, bool) {
	for _, facter := range m.facters {
		v, found := facter.Fact(name)
		if found {
			return v, true
		}
	}
	return "", false
}
