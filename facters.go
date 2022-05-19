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
