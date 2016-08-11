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
	"os"

	"github.com/spf13/cobra"
)

func NewKubeadmCommand(f *cmdutil.Factory, in io.Reader, out, err io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "kubeadm",
		Short: "kubeadm: bootstrap a kubernetes cluster.",
		Long:  `kubeadm: bootstrap a kubernetes cluster.`,
	}

	cmds.AddCommand(NewCmdInit(out))
	cmds.AddCommand(NewCmdJoin(out))
	cmds.AddCommand(NewCmdUser(out))

	cmds.PersistentFlags().StringVarP(&config.disco, "discovery", "", "gossip",
		`which discovery mechanism to use for cluster bootstrap (choose from "gossip", "out-of-band").`)

	return cmds
}
