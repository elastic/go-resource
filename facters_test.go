package resource

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvFacter(t *testing.T) {
	facter := &EnvFacter{}
	facter.Prefix = "TEST"

	manager := NewManager()
	manager.AddFacter(facter)

	fact := "testfact"
	expectedValue := "somevalue"

	value, found := manager.Fact(fact)
	assert.Equal(t, "", value)
	assert.False(t, found)

	os.Setenv(facter.Prefix+"_"+fact, expectedValue)

	value, found = manager.Fact(fact)
	assert.Equal(t, expectedValue, value)
	assert.True(t, found)
}
