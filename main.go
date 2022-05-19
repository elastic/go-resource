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
