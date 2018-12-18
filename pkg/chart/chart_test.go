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

package chart

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/helm/chart-testing/pkg/util"

	"github.com/helm/chart-testing/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeGit struct{}

func (g fakeGit) FileExistsOnBranch(file string, remote string, branch string) bool {
	return true
}

func (g fakeGit) Show(file string, remote string, branch string) (string, error) {
	return "", nil
}

func (g fakeGit) MergeBase(commit1 string, commit2 string) (string, error) {
	return "", nil
}

func (g fakeGit) ListChangedFilesInDirs(commit string, dirs ...string) ([]string, error) {
	return []string{
		"incubator/excluded/Chart.yaml",
		"incubator/excluded/values.yaml",
		"incubator/bar/README.md",
		"incubator/bar/README.md",
		"incubator/excluded/templates/configmap.yaml",
		"incubator/excluded/values.yaml",
		"stable/blah/Chart.yaml",
		"stable/blah/README.md",
		"stable/this-is-no-chart-dir/foo.md",
	}, nil
}

func (g fakeGit) GetUrlForRemote(remote string) (string, error) {
	return "git@github.com/helm/chart-testing", nil
}

type fakeDirLister struct{}

func (l fakeDirLister) ListChildDirs(parentDir string, test func(dir string) bool) ([]string, error) {
	if parentDir == "stable" {
		var dirs []string
		for _, dir := range []string{"stable/foo", "stable/excluded"} {
			if test(dir) {
				dirs = append(dirs, dir)
			}
		}
		return dirs, nil
	}
	return []string{"incubator/bar"}, nil
}

type fakeChartUtils struct{}

func (v fakeChartUtils) IsChartDir(dir string) bool {
	return dir != "stable/this-is-no-chart-dir"
}

func (v fakeChartUtils) ReadChartYaml(dir string) (*util.ChartYaml, error) {
	chartUtils := util.ChartUtils{}
	return chartUtils.ReadChartYaml(dir)
}

type fakeAccountValidator struct{}

func (v fakeAccountValidator) Validate(repoDomain string, account string) error {
	if strings.HasPrefix(account, "valid") {
		return nil
	}
	return errors.New(fmt.Sprintf("Error validating account: %s", account))
}

type fakeLinter struct{}

func (l fakeLinter) YamlLint(yamlFile, configFile string) error { return nil }
func (l fakeLinter) Yamale(yamlFile, schemaFile string) error   { return nil }

type fakeLinter2 struct {
	mock.Mock
}

func (l *fakeLinter2) YamlLint(yamlFile, configFile string) error {
	l.Called(yamlFile, configFile)
	return nil
}
func (l *fakeLinter2) Yamale(yamlFile, schemaFile string) error {
	l.Called(yamlFile, schemaFile)
	return nil
}

type fakeHelm struct{}

func (h fakeHelm) Init() error                                          { return nil }
func (h fakeHelm) AddRepo(name, url string) error                       { return nil }
func (h fakeHelm) BuildDependencies(chart string) error                 { return nil }
func (h fakeHelm) LintWithValues(chart string, valuesFile string) error { return nil }
func (h fakeHelm) InstallWithValues(chart string, valuesFile string, namespace string, release string) error {
	return nil
}
func (h fakeHelm) Test(release string) error {
	return nil
}
func (h fakeHelm) DeleteRelease(release string) {}

var ct Testing

func init() {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"stable", "incubator"},
	}
	ct = Testing{
		config:           cfg,
		directoryLister:  fakeDirLister{},
		git:              fakeGit{},
		chartUtils:       fakeChartUtils{},
		accountValidator: fakeAccountValidator{},
		linter:           fakeLinter{},
		helm:             fakeHelm{},
	}
}

func TestComputeChangedChartDirectories(t *testing.T) {
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"incubator/bar", "stable/blah"}
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestReadAllChartDirectories(t *testing.T) {
	actual, err := ct.ReadAllChartDirectories()
	expected := []string{"stable/foo", "incubator/bar"}
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestValidateMaintainers(t *testing.T) {
	var testDataSlice = []struct {
		name     string
		chartDir string
		expected bool
	}{
		{"valid", "testdata/valid_maintainers", true},
		{"invalid", "testdata/invalid_maintainers", false},
		{"no-maintainers", "testdata/no_maintainers", false},
		{"empty-maintainers", "testdata/empty_maintainers", false},
		{"valid-deprecated", "testdata/valid_maintainers_deprecated", false},
		{"no-maintainers-deprecated", "testdata/no_maintainers_deprecated", true},
	}

	for _, testData := range testDataSlice {
		t.Run(testData.name, func(t *testing.T) {
			err := ct.ValidateMaintainers(testData.chartDir)
			assert.Equal(t, testData.expected, err == nil)
		})
	}
}

func TestLintChartMaintainerValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(validate bool) {
		ct.config.ValidateMaintainers = validate

		var suffix string
		if validate {
			suffix = "with-validation"
		} else {
			suffix = "without-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("maintainers-%s", suffix), "testdata/valid_maintainers", true},
			{fmt.Sprintf("no-maintainers-%s", suffix), "testdata/no_maintainers", !validate},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				result := ct.LintChart(testData.chartDir, []string{})
				assert.Equal(t, testData.expected, result.Error == nil)
			})
		}
	}

	runTests(true)
	runTests(false)
}

func TestLintNoChartSchemaValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(noValidation bool, callsYamlLint, callsYamale int) {
		var fakeMockLinter = new(fakeLinter2)

		fakeMockLinter.On("Yamale", mock.Anything, mock.Anything).Return(true)
		fakeMockLinter.On("YamlLint", mock.Anything, mock.Anything).Return(true)

		ct.linter = fakeMockLinter
		ct.config.NoChartSchemaValidation = noValidation
		ct.config.ValidateMaintainers = false
		ct.config.NoYamlLint = false

		var suffix string
		if noValidation {
			suffix = "without-validation"
		} else {
			suffix = "with-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("schema-%s", suffix), "testdata/test_lints", true},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				result := ct.LintChart(testData.chartDir, []string{})
				assert.Equal(t, testData.expected, result.Error == nil)
				fakeMockLinter.AssertNumberOfCalls(t, "Yamale", callsYamale)
				fakeMockLinter.AssertNumberOfCalls(t, "YamlLint", callsYamlLint)
			})
		}
	}

	// will run the schema validation
	runTests(false, 2, 1)
	// will not run the schema validation
	runTests(true, 2, 0)

}

func TestLintNoYamlLintValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(noValidation bool, callsYamlLint, callsYamale int) {

		var fakeMockLinter = new(fakeLinter2)

		fakeMockLinter.On("Yamale", mock.Anything, mock.Anything).Return(true)
		fakeMockLinter.On("YamlLint", mock.Anything, mock.Anything).Return(true)

		ct.linter = fakeMockLinter
		ct.config.NoYamlLint = noValidation
		ct.config.NoChartSchemaValidation = false
		ct.config.ValidateMaintainers = false

		var suffix string
		if noValidation {
			suffix = "without-yaml-validation"
		} else {
			suffix = "with-yaml-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("lint-%s", suffix), "testdata/test_lints", true},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				result := ct.LintChart(testData.chartDir, []string{})
				assert.Equal(t, testData.expected, result.Error == nil)
				fakeMockLinter.AssertNumberOfCalls(t, "Yamale", callsYamale)
				fakeMockLinter.AssertNumberOfCalls(t, "YamlLint", callsYamlLint)
			})
		}
	}

	// will run the lint validation
	runTests(false, 2, 1)
	// will not run the lint validation
	runTests(true, 0, 1)
}
