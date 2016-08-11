/*
Copyright 2016 The Kubernetes Authors.

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

package app

import (
	"os"

	"k8s.io/kubernetes/pkg/kubeadm/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

// TODO add this to hyperkube?
func Run() error {
	cmd := cmd.NewKubeadmCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	return cmd.Execute()
}
