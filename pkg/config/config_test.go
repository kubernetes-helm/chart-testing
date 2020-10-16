// Copyright The Helm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalYaml(t *testing.T) {
	loadAndAssertConfigFromFile(t, "test_config.yaml")
}

func TestUnmarshalJson(t *testing.T) {
	loadAndAssertConfigFromFile(t, "test_config.json")
}

func loadAndAssertConfigFromFile(t *testing.T, configFile string) {
	cfg, _ := LoadConfiguration(configFile, &cobra.Command{
		Use: "install",
	}, true)

	require.Equal(t, "origin", cfg.Remote)
	require.Equal(t, "master", cfg.TargetBranch)
	require.Equal(t, "pr-42", cfg.BuildId)
	require.Equal(t, "my-lint-conf.yaml", cfg.LintConf)
	require.Equal(t, "my-chart-yaml-schema.yaml", cfg.ChartYamlSchema)
	require.Equal(t, true, cfg.ValidateMaintainers)
	require.Equal(t, true, cfg.ValidateChartSchema)
	require.Equal(t, true, cfg.ValidateYaml)
	require.Equal(t, true, cfg.StrictLint)
	require.Equal(t, true, cfg.CheckVersionIncrement)
	require.Equal(t, false, cfg.ProcessAllCharts)
	require.Equal(t, []string{"incubator=https://incubator"}, cfg.ChartRepos)
	require.Equal(t, []string{"incubator=--username test"}, cfg.HelmRepoExtraArgs)
	require.Equal(t, []string{"stable", "incubator"}, cfg.ChartDirs)
	require.Equal(t, []string{"common"}, cfg.ExcludedCharts)
	require.Equal(t, "--timeout 300", cfg.HelmExtraArgs)
	require.Equal(t, true, cfg.Upgrade)
	require.Equal(t, true, cfg.SkipMissingValues)
	require.Equal(t, "default", cfg.Namespace)
	require.Equal(t, "release", cfg.ReleaseLabel)
}
