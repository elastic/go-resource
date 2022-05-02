package resource

import (
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
