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

package kubeadm

import (
	"io"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubelet"
)

type BootstrapParams struct {
	// A struct with methods that implement Start() and Discover()
	// kubeadm will persist this struct to disk and kubelet will read it.
	Discovery kubelet.Discovery
}

func NewKubeadmCommand(f *cmdutil.Factory, in io.Reader, out, err io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "kubeadm",
		Short: "kubeadm: bootstrap a secure kubernetes cluster easily.",
		Long: `kubeadm: bootstrap a secure kubernetes cluster easily.

	/==========================================================\
	| KUBEADM IS ALPHA, DO NOT USE IT FOR PRODUCTION CLUSTERS! |
	|                                                          |
	| But, please try it out! Give us feedback at:             |
	| https://github.com/kubernetes/kubernetes/issues          |
	\==========================================================/

Example usage:

	Create a two-server cluster with one master (which controls the cluster),
	and one node (where workloads, like pods and containers run).

	master# kubeadm init master
	Your token is: <token>

	node# kubeadm join node --token=<token> <ip-of-master>
`,
	}
	// TODO also print the alpha warning when running any commands, as well as
	// in the help text.

	bootstrapParams := &BootstrapParams{}
	cmds.AddCommand(NewCmdInit(out, bootstrapParams))
	cmds.AddCommand(NewCmdJoin(out, bootstrapParams))
	cmds.AddCommand(NewCmdUser(out, bootstrapParams))
	cmds.AddCommand(NewCmdManual(out, bootstrapParams))

	return cmds
}
