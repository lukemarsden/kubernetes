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
	cmd.AddCommand(NewCmdManualBootstrapMaster(out, params))
	cmd.AddCommand(NewCmdManualBootstrapNode(out, params))

	return cmd
}

func NewCmdManualBootstrapMaster(out io.Writer, params *BootstrapParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manually bootstrap a master 'out-of-band'",
		Long: `Manually bootstrap a master 'out-of-band'.

Will create TLS certificates and set up static pods for Kubernetes master
components.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			/*

				api server & controller manager
				===============================
				* service_cluster_ip_range - can default to 10.16.0.0/12

				* cloud provider - "fake"
				* cluster name - "kubernetes"
				* kubernetes version (for container images) - can deduce?
				* docker registry, image name ("hyperkube") - can have defaults
				* secure port - default to 443

			*/
			err := writeStaticPodsOnMaster()
			if err != nil {
				return err
			}
			err = generateAndWriteCertificatesOnMaster(params.ApiServerDNSName)
			if err != nil {
				return err
			}
			out.Write([]byte(`CA cert is written to XXX. Please scp this to all your nodes before running
'kubeadm manual bootstrap node --ca-cert-file <path-to-ca-cert>
--api-server-urls http://<ip-of-master>:8080/`))
			err := writeParamsIfNotExists(params)
			if err != nil {
				out.Write([]byte(fmt.Sprintf("Unable to write config for master:\n%s\n", err)))
			}
		},
	}
	var discovery *kubelet.OutOfBandDiscovery
	discovery = &kubelet.OutOfBandDiscovery{}
	params.Discovery = discovery

	cmd.PersistentFlags().StringVarP(&discovery.ApiServerDNSName, "api-dns-name", "", "",
		`(optional) DNS name for the API server, will be encoded into
            subjectAltName in the resulting (generated) TLS certificates`)

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
	var discovery *kubelet.OutOfBandDiscovery
	discovery = &kubelet.OutOfBandDiscovery{}
	params.Discovery = discovery

	cmd.PersistentFlags().StringVarP(&discovery.CaCertFile, "ca-cert-file", "", "",
		`Path to a CA cert file in PEM format. The same CA cert must be distributed to
            all servers.`)
	cmd.PersistentFlags().StringVarP(&discovery.ApiServerURLs, "api-server-urls", "", "",
		`Comma separated list of API server URLs. Typically this might be just
            https://<address-of-master>:8080/`)

	return cmd
}
