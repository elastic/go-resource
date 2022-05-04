package resource

import (
	"log"
)

// Main is a helper to generate single binaries to manage a collection of resources.
type Main struct {
	// Facters is the list of facters used by this command.
	Facters []Facter

	// Providers is the list of providers used by this command.
	Providers map[string]Provider

	// Resources is the list of resources managed by this command.
	Resources Resources
}

func (c *Main) Run() error {
	manager := NewManager()

	for name, provider := range c.Providers {
		manager.RegisterProvider(name, provider)
	}

	for _, facter := range c.Facters {
		manager.AddFacter(facter)
	}

	results, err := manager.Apply(c.Resources)
	for _, result := range results {
		log.Println(result)
	}
	return err
}
