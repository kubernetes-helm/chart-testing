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

package tool

import (
	"fmt"

	"github.com/helm/chart-testing/v3/pkg/exec"
)

type Helm struct {
	exec      exec.ProcessExecutor
	extraArgs []string
	lintArgs  []string
}

func NewHelm(exec exec.ProcessExecutor, extraArgs []string, lintArgs []string) Helm {
	return Helm{
		exec:      exec,
		extraArgs: extraArgs,
		lintArgs:  lintArgs,
	}
}

func (h Helm) AddRepo(name string, url string, extraArgs []string) error {
	return h.exec.RunProcess("helm", "repo", "add", name, url, extraArgs)
}

func (h Helm) BuildDependencies(chart string) error {
	return h.exec.RunProcess("helm", "dependency", "build", chart)
}

func (h Helm) LintWithValues(chart string, valuesFile string) error {
	var values []string
	if valuesFile != "" {
		values = []string{"--values", valuesFile}
	}

	return h.exec.RunProcess("helm", "lint", h.lintArgs, chart, values)
}

func (h Helm) InstallWithValues(chart string, valuesFile string, namespace string, release string) error {
	var values []string
	if valuesFile != "" {
		values = []string{"--values", valuesFile}
	}

	if err := h.exec.RunProcess("helm", "install", release, chart, "--namespace", namespace,
		"--wait", values, h.extraArgs); err != nil {
		return err
	}

	return nil
}

func (h Helm) Upgrade(chart string, namespace string, release string) error {
	if err := h.exec.RunProcess("helm", "upgrade", release, chart, "--namespace", namespace,
		"--reuse-values", "--wait", h.extraArgs); err != nil {
		return err
	}

	return nil
}

func (h Helm) Test(namespace string, release string) error {
	return h.exec.RunProcess("helm", "test", release, "--namespace", namespace, h.extraArgs)
}

func (h Helm) DeleteRelease(namespace string, release string) {
	fmt.Printf("Deleting release '%s'...\n", release)
	if err := h.exec.RunProcess("helm", "uninstall", release, "--namespace", namespace, h.extraArgs); err != nil {
		fmt.Println("Error deleting Helm release:", err)
	}
}

func (h Helm) Version() (string, error) {
	return h.exec.RunProcessAndCaptureOutput("helm", "version", "--short")
}
