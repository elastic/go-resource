package resource

import "os"

const defaultEnvFacterPrefix = "FACT"

// EnvFacter is a facter that gets facts from environment variables.
// Facts can be defined in environment variables starting with the "FACT"
// prefix.
// For example the "runtime" fact could be set with "FACT_runtime".
// A different setting can be selected using the Prefix attribute.
type EnvFacter struct {
	// Prefix used to find facts in environment variables. If not
	// set, "FACT" is used.
	Prefix string
}

// Fact returns the value of a fact obtained from the environment if it
// exists. If not, it returns an empty string and false.
func (f *EnvFacter) Fact(name string) (string, bool) {
	prefix := f.Prefix
	if prefix == "" {
		prefix = defaultEnvFacterPrefix
	}
	envName := prefix + "_" + name
	return os.LookupEnv(envName)
}
