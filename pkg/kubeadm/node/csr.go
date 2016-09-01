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

package kubenode

import (
	"fmt"
	"io/ioutil"
	"strings"

	kubeletapp "k8s.io/kubernetes/cmd/kubelet/app"
	unversionedcertificates "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/certificates/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
	kubeadmutil "k8s.io/kubernetes/pkg/kubeadm/util"
	utilcertificates "k8s.io/kubernetes/pkg/util/certificates"
)

func getNodeName() string {
	return "TODO"
}

func PerformTLSBootstrapFromParams(params *kubeadmapi.BootstrapParams) (*clientcmdapi.Config, error) {
	caCert, err := ioutil.ReadFile(params.Discovery.CaCertFile)
	if err != nil {
		return nil, err
	}

	return PerformTLSBootstrap(params, strings.Split(params.Discovery.ApiServerURLs, ",")[0], caCert)
}

// Create a restful client for doing the certificate signing request.
func PerformTLSBootstrap(params *kubeadmapi.BootstrapParams, apiEndpoint string, caCert []byte) (*clientcmdapi.Config, error) {
	// TODO try all the api servers until we find one that works
	bareClientConfig := kubeadmutil.CreateBasicClientConfig("kubernetes", apiEndpoint, caCert)

	nodeName := getNodeName()

	bootstrapClientConfig, err := clientcmd.NewDefaultClientConfig(
		*kubeadmutil.MakeClientConfigWithToken(
			bareClientConfig, "kubernetes", fmt.Sprintf("kubelet-%s", nodeName), params.Discovery.BearerToken,
		),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("<node/csr> failed to create an API client configuration [%s]", err)
	}
	//fmt.Printf("loaded bootstrap client configuration: %#v\n", bootstrapClientConfig)

	//fmt.Println("creating CSR...")
	client, err := unversionedcertificates.NewForConfig(bootstrapClientConfig)
	if err != nil {
		return nil, fmt.Errorf("<node/csr> failed to create an API client [%s]", err)
	}
	csrClient := client.CertificateSigningRequests()

	keyData, err := utilcertificates.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("<node/csr> failed to generating private key [%s]", err)
	}
	//fmt.Println("CSR created, asking the API server to sign it...")
	// Pass 'requestClientCertificate()' the CSR client, existing key data, and node name to
	// request for client certificate from the API server.
	certData, err := kubeletapp.RequestClientCertificate(csrClient, keyData, nodeName)
	if err != nil {
		return nil, fmt.Errorf("<node/csr> failed to request signed certificate from the API server [%s]", err)
	}

	finalConfig := kubeadmutil.MakeClientConfigWithCerts(
		bareClientConfig, "kubernetes", fmt.Sprintf("kubelet-%s", nodeName),
		keyData, certData,
	)

	return finalConfig, nil
}
