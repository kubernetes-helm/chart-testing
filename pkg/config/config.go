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
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/helm/chart-testing/v3/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	homeDir, _            = homedir.Dir()
	configSearchLocations = []string{
		".",
		path.Join(homeDir, ".ct"),
		"/usr/local/etc/ct",
		"/etc/ct",
	}
)

type Configuration struct {
	Remote                string   `mapstructure:"remote"`
	TargetBranch          string   `mapstructure:"target-branch"`
	Since                 string   `mapstructure:"since"`
	BuildId               string   `mapstructure:"build-id"`
	LintConf              string   `mapstructure:"lint-conf"`
	ChartYamlSchema       string   `mapstructure:"chart-yaml-schema"`
	ValidateMaintainers   bool     `mapstructure:"validate-maintainers"`
	ValidateChartSchema   bool     `mapstructure:"validate-chart-schema"`
	ValidateYaml          bool     `mapstructure:"validate-yaml"`
	StrictLint            bool     `mapstructure:"strict-lint"`
	CheckVersionIncrement bool     `mapstructure:"check-version-increment"`
	ProcessAllCharts      bool     `mapstructure:"all"`
	Charts                []string `mapstructure:"charts"`
	ChartRepos            []string `mapstructure:"chart-repos"`
	ChartDirs             []string `mapstructure:"chart-dirs"`
	ExcludedCharts        []string `mapstructure:"excluded-charts"`
	HelmExtraArgs         string   `mapstructure:"helm-extra-args"`
	HelmRepoExtraArgs     []string `mapstructure:"helm-repo-extra-args"`
	Debug                 bool     `mapstructure:"debug"`
	Upgrade               bool     `mapstructure:"upgrade"`
	SkipMissingValues     bool     `mapstructure:"skip-missing-values"`
	Namespace             string   `mapstructure:"namespace"`
	ReleaseLabel          string   `mapstructure:"release-label"`
}

func LoadConfiguration(cfgFile string, cmd *cobra.Command, printConfig bool) (*Configuration, error) {
	v := viper.New()

	cmd.Flags().VisitAll(func(flag *flag.Flag) {
		flagName := flag.Name
		if flagName != "config" && flagName != "help" {
			if err := v.BindPFlag(flagName, flag); err != nil {
				// can't really happen
				panic(fmt.Sprintln(errors.Wrapf(err, "Error binding flag '%s'", flagName)))
			}
		}
	})

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.SetEnvPrefix("CT")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("ct")
		for _, searchLocation := range configSearchLocations {
			v.AddConfigPath(searchLocation)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// Only error out for specified config file. Ignore for default locations.
			return nil, errors.Wrap(err, "Error loading config file")
		}
	} else {
		if printConfig {
			fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed())
		}
	}

	isLint := strings.Contains(cmd.Use, "lint")
	isInstall := strings.Contains(cmd.Use, "install")

	cfg := &Configuration{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling configuration")
	}

	if cfg.ProcessAllCharts && len(cfg.Charts) > 0 {
		return nil, errors.New("specifying both, '--all' and '--charts', is not allowed")
	}

	if cfg.Namespace != "" && cfg.ReleaseLabel == "" {
		return nil, errors.New("specifying '--namespace' without '--release-label' is not allowed")
	}

	// Disable upgrade (this does some expensive dependency building on previous revisions)
	// when neither "install" nor "lint-and-install" have not been specified.
	cfg.Upgrade = isInstall && cfg.Upgrade
	if (cfg.TargetBranch == "" || cfg.Remote == "") && cfg.Upgrade {
		return nil, errors.New("specifying '--upgrade=true' without '--target-branch' or '--remote', is not allowed")
	}

	chartYamlSchemaPath := cfg.ChartYamlSchema
	if chartYamlSchemaPath == "" {
		var err error
		cfgFile, err = findConfigFile("chart_schema.yaml")
		if err != nil && isLint && cfg.ValidateChartSchema {
			return nil, errors.New("'chart_schema.yaml' neither specified nor found in default locations")
		}
		cfg.ChartYamlSchema = cfgFile
	}

	lintConfPath := cfg.LintConf
	if lintConfPath == "" {
		var err error
		cfgFile, err = findConfigFile("lintconf.yaml")
		if err != nil && isLint && cfg.ValidateYaml {
			return nil, errors.New("'lintconf.yaml' neither specified nor found in default locations")
		}
		cfg.LintConf = cfgFile
	}

	if len(cfg.Charts) > 0 || cfg.ProcessAllCharts {
		fmt.Fprintln(os.Stderr, "Version increment checking disabled.")
		cfg.CheckVersionIncrement = false
	}

	if printConfig {
		printCfg(cfg)
	}

	return cfg, nil
}

func printCfg(cfg *Configuration) {
	util.PrintDelimiterLineToWriter(os.Stderr, "-")
	fmt.Fprintln(os.Stderr, " Configuration")
	util.PrintDelimiterLineToWriter(os.Stderr, "-")

	e := reflect.ValueOf(cfg).Elem()
	typeOfCfg := e.Type()

	for i := 0; i < e.NumField(); i++ {
		var pattern string
		switch e.Field(i).Kind() {
		case reflect.Bool:
			pattern = "%s: %t\n"
		default:
			pattern = "%s: %s\n"
		}
		fmt.Fprintf(os.Stderr, pattern, typeOfCfg.Field(i).Name, e.Field(i).Interface())
	}

	util.PrintDelimiterLineToWriter(os.Stderr, "-")
}

func findConfigFile(fileName string) (string, error) {
	for _, location := range configSearchLocations {
		filePath := path.Join(location, fileName)
		if util.FileExists(filePath) {
			return filePath, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Config file not found: %s", fileName))
}
