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
	"github.com/helm/chart-testing/pkg/exec"
)

type Helm struct {
	exec      exec.ProcessExecutor
	kubectl   Kubectl
	extraArgs []string
}

func NewHelm(kubectl Kubectl, extraArgs []string) Helm {
	return Helm{
		exec:      exec.ProcessExecutor{},
		kubectl:   kubectl,
		extraArgs: extraArgs,
	}
}

func (h Helm) Init() error {
	return h.exec.RunProcess("helm", "init", "--client-only")
}

func (h Helm) AddRepo(name string, url string) error {
	return h.exec.RunProcess("helm", "repo", "add", name, url)
}

func (h Helm) BuildDependencies(chart string) error {
	return h.exec.RunProcess("helm", "dependency", "build", chart)
}

func (h Helm) Lint(chart string) error {
	return h.exec.RunProcess("helm", "lint", chart)
}

func (h Helm) LintWithValues(chart string, valuesFile string) error {
	return h.exec.RunProcess("helm", "lint", chart, "--values", valuesFile)
}

func (h Helm) Install(chart string, namespace string, release string) error {
	return h.InstallWithValues(chart, "", namespace, release)
}

func (h Helm) InstallWithValues(chart string, valuesFile string, namespace string, release string) error {
	var values []string
	if valuesFile != "" {
		values = []string{"--values", valuesFile}
	}

	if err := h.exec.RunProcess("helm", "install", chart, "--name", release, "--namespace", namespace,
		"--wait", values, h.extraArgs); err != nil {
		return err
	}

	if err := h.exec.RunProcess("helm", "test", release, h.extraArgs); err != nil {
		return err
	}

	return h.kubectl.WaitForDeployments(namespace)
}

func (h Helm) DeleteRelease(release string) {
	fmt.Printf("Deleting release '%s'...\n", release)
	if err := h.exec.RunProcess("helm", "delete", "--purge", release, h.extraArgs); err != nil {
		fmt.Println("Error deleting Helm release:", err)
	}
}
