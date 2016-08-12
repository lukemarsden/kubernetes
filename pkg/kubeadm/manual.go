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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/kubelet"
)

func NewCmdManual(out io.Writer, params *BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual",
		Short: "Advanced, less-automated functionality, for power users.",
		// TODO put example usage in the Long description here
	}
	cmd.AddCommand(NewCmdManualBootstrap(out, params))
	return cmd
}

func NewCmdManualBootstrap(out io.Writer, params *BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Manually bootstrap a cluster 'out-of-band'",
		Long: `Manually bootstrap a cluster 'out-of-band', by generating and distributing a CA
certificate to all your servers and specifying and (list of) API server URLs.`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	var discovery *kubelet.OutOfBandDiscovery
	discovery = &kubelet.OutOfBandDiscovery{}
	params.Discovery = discovery

	cmd.PersistentFlags().StringVarP(&discovery.CaCertFile, "cacertfile", "", "",
		`Path to a CA cert file in PEM format. The same CA cert must be distributed to
            all servers.`)
	cmd.PersistentFlags().StringVarP(&discovery.ApiServerURLs, "apiserverurls", "", "",
		`Comma separated list of API server URLs. Typically this might be just
            https://<address-of-master>:8080/`)

	cmd.AddCommand(NewCmdManualBootstrapMaster(out, bootstrapParams))
	cmd.AddCommand(NewCmdManualBootstrapNode(out, bootstrapParams))

	return cmd
}

func NewCmdManualBootstrapMaster(out io.Writer, params *BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manually bootstrap a node 'out-of-band'",

		Run: func(cmd *cobra.Command, args []string) {
			err := writeParamsIfNotExists(params)
			if err != nil {
				out.Write([]byte(fmt.Sprintf("Unable to write config for master:\n%s\n", err)))
			}
		},
	}
	return cmd
}

func NewCmdManualBootstrapNode(out io.Writer, params *BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manually bootstrap a node 'out-of-band'",

		Run: func(cmd *cobra.Command, args []string) {
			err := writeParamsIfNotExists(params)
			if err != nil {
				out.Write([]byte(fmt.Sprintf("Unable to write config for node:\n%s\n", err)))
			}
		},
	}
	return cmd
}
