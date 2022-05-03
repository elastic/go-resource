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
