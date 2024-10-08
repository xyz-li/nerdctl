/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"errors"
	"testing"

	"github.com/containerd/containerd/v2/defaults"

	"github.com/containerd/nerdctl/v2/pkg/testutil"
	"github.com/containerd/nerdctl/v2/pkg/testutil/nerdtest"
	"github.com/containerd/nerdctl/v2/pkg/testutil/test"
)

func TestMain(m *testing.M) {
	testutil.M(m)
}

// TestUnknownCommand tests https://github.com/containerd/nerdctl/issues/487
func TestUnknownCommand(t *testing.T) {
	nerdtest.Setup()

	var unknownSubCommand = errors.New("unknown subcommand")

	testGroup := &test.Group{
		{
			Description: "non-existent-command",
			Command:     test.RunCommand("non-existent-command"),
			Expected:    test.Expects(1, []error{unknownSubCommand}, nil),
		},
		{
			Description: "non-existent-command info",
			Command:     test.RunCommand("non-existent-command", "info"),
			Expected:    test.Expects(1, []error{unknownSubCommand}, nil),
		},
		{
			Description: "system non-existent-command",
			Command:     test.RunCommand("system", "non-existent-command"),
			Expected:    test.Expects(1, []error{unknownSubCommand}, nil),
		},
		{
			Description: "system non-existent-command info",
			Command:     test.RunCommand("system", "non-existent-command", "info"),
			Expected:    test.Expects(1, []error{unknownSubCommand}, nil),
		},
		{
			Description: "system",
			Command:     test.RunCommand("system"),
			Expected:    test.Expects(0, nil, nil),
		},
		{
			Description: "system info",
			Command:     test.RunCommand("system", "info"),
			Expected:    test.Expects(0, nil, test.Contains("Kernel Version:")),
		},
		{
			Description: "info",
			Command:     test.RunCommand("info"),
			Expected:    test.Expects(0, nil, test.Contains("Kernel Version:")),
		},
	}

	testGroup.Run(t)
}

// TestNerdctlConfig validates the configuration precedence [CLI, Env, TOML, Default] and broken config rejection
func TestNerdctlConfig(t *testing.T) {
	nerdtest.Setup()

	tc := &test.Case{
		Description: "Nerdctl configuration",
		// Docker does not support nerdctl.toml obviously
		Require: test.Not(nerdtest.Docker),
		SubTests: []*test.Case{
			{
				Description: "Default",
				Command:     test.RunCommand("info", "-f", "{{.Driver}}"),
				Expected:    test.Expects(0, nil, test.Equals(defaults.DefaultSnapshotter+"\n")),
			},
			{
				Description: "TOML > Default",
				Command:     test.RunCommand("info", "-f", "{{.Driver}}"),
				Expected:    test.Expects(0, nil, test.Equals("dummy-snapshotter-via-toml\n")),
				Data:        test.WithConfig(nerdtest.NerdctlToml, `snapshotter = "dummy-snapshotter-via-toml"`),
			},
			{
				Description: "Cli > TOML > Default",
				Command:     test.RunCommand("info", "-f", "{{.Driver}}", "--snapshotter=dummy-snapshotter-via-cli"),
				Expected:    test.Expects(0, nil, test.Equals("dummy-snapshotter-via-cli\n")),
				Data:        test.WithConfig(nerdtest.NerdctlToml, `snapshotter = "dummy-snapshotter-via-toml"`),
			},
			{
				Description: "Env > TOML > Default",
				Command:     test.RunCommand("info", "-f", "{{.Driver}}"),
				Env:         map[string]string{"CONTAINERD_SNAPSHOTTER": "dummy-snapshotter-via-env"},
				Expected:    test.Expects(0, nil, test.Equals("dummy-snapshotter-via-env\n")),
				Data:        test.WithConfig(nerdtest.NerdctlToml, `snapshotter = "dummy-snapshotter-via-toml"`),
			},
			{
				Description: "Cli > Env > TOML > Default",
				Command:     test.RunCommand("info", "-f", "{{.Driver}}", "--snapshotter=dummy-snapshotter-via-cli"),
				Env:         map[string]string{"CONTAINERD_SNAPSHOTTER": "dummy-snapshotter-via-env"},
				Expected:    test.Expects(0, nil, test.Equals("dummy-snapshotter-via-cli\n")),
				Data:        test.WithConfig(nerdtest.NerdctlToml, `snapshotter = "dummy-snapshotter-via-toml"`),
			},
			{
				Description: "Broken config",
				Command:     test.RunCommand("info"),
				Expected:    test.Expects(1, []error{errors.New("failed to load nerdctl config")}, nil),
				Data: test.WithConfig(nerdtest.NerdctlToml, `# containerd config, not nerdctl config
version = 2`),
			},
		},
	}

	tc.Run(t)
}
