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
package kubemaster

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
)

func createKubeProxyPodSpec(params *kubeadmapi.BootstrapParams) api.PodSpec {
	// TODO does it need API credentials or it will use default service account?
	// we have variely of options here... we can even do a CSR dance, or just
	// reuse kubelet.conf and mount it either from host path or as a secret...
	privilegedTrue := true
	return api.PodSpec{
		SecurityContext: &api.PodSecurityContext{HostNetwork: true},
		Containers: []api.Container{{
			Name:  "kube-proxy",
			Image: params.EnvParams["hyperkube_image"],
			Command: []string{
				"/hyperkube",
				"proxy",
				COMPONENT_LOGLEVEL,
			},
			SecurityContext: &api.SecurityContext{Privileged: &privilegedTrue},
		}},
	}
}

func CreateEssentialAddons(params *kubeadmapi.BootstrapParams, client *clientset.Clientset) error {
	kubeProxyDaemonSet := NewDaemonSet("kube-proxy", createKubeProxyPodSpec(params))

	if _, err := client.Extensions().DaemonSets(api.NamespaceSystem).Create(kubeProxyDaemonSet); err != nil {
		return fmt.Errorf("failed creating kube-proxy addon [%s]", err)
	}

	return nil
}
