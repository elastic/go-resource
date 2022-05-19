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

import "fmt"

type Migration func(*Manager) (ApplyResults, error)

type Migrator struct {
	version    Versioner
	migrations []migrationEntry
}

type Versioner interface {
	Current() uint
	Set(uint) error
}

type migrationEntry struct {
	version   uint
	migration Migration
}

func NewMigrator(versioner Versioner) *Migrator {
	return &Migrator{version: versioner}
}

func (m *Migrator) AddMigration(version uint, migration Migration) {
	if len(m.migrations) > 0 && version <= m.migrations[len(m.migrations)-1].version {
		panic("adding migration for a smaller version")
	}
	m.migrations = append(m.migrations, migrationEntry{
		version:   version,
		migration: migration,
	})
}

func (m *Migrator) RunMigrations(manager *Manager) (ApplyResults, error) {
	currentVersion := m.version.Current()
	var results ApplyResults
	for _, entry := range m.migrations {
		if entry.version <= currentVersion {
			continue
		}

		r, err := entry.migration(manager)
		results = append(results, r...)
		if err != nil {
			return results, err
		}

		err = m.version.Set(entry.version)
		if err != nil {
			return results, fmt.Errorf("failed to save migration version: %w", err)
		}
	}

	return results, nil
}
