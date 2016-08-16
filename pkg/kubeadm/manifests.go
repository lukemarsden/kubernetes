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
	"bytes"
	"encoding/json"
	_ "fmt"
	"os"
	"path"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// Static pod definitions in golang form are included below so that `kubeadm
// init master` and `kubeadm manual bootstrap master` can get going.

const (
	COMPONENT_LOGLEVEL       = "--v=4"
	SERVICE_CLUSTER_IP_RANGE = "--service-cluster-ip-range=10.16.0.0/12"
	CLUSTER_NAME             = "--cluster-name=kubernetes"
	MASTER                   = "--master=127.0.0.1:8080"
	HYPERKUBE_IMAGE          = "errordeveloper/hyperquick:master"
)

func writeStaticPodManifests(params *BootstrapParams) error {
	staticPodSpecs := map[string]api.Pod{
		// TODO this needs a volume
		"etcd": componentPod(api.Container{
			Command: []string{
				"/bin/sh", // TODO do we really need the parent shell here?
				"-c",
				"/usr/local/bin/etcd --listen-peer-urls http://127.0.0.1:2380 --addr 127.0.0.1:2379 --bind-addr 127.0.0.1:2379 --data-dir /var/etcd/data",
			},
			Image:         "gcr.io/google_containers/etcd:2.2.1", // TODO parametrise
			LivenessProbe: componentProbe(2379, "/health"),
			Name:          "etcd-server",
			Ports: []api.ContainerPort{
				{Name: "serverport", ContainerPort: 2380, HostPort: 2380},
				{Name: "clientport", ContainerPort: 2379, HostPort: 2379},
			},
			Resources: componentResources("200m"),
		}),
		// TODO bind-mount certs in
		"kube-apiserver": componentPod(api.Container{
			Name:  "kube-apiserver",
			Image: HYPERKUBE_IMAGE,
			Command: []string{
				"/hyperkube",
				"apiserver",
				"--address=127.0.0.1",
				"--etcd-servers=http://127.0.0.1:2379",
				"--cloud-provider=fake", // TODO parametrise
				"--admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,ResourceQuota",
				SERVICE_CLUSTER_IP_RANGE,
				"--service-account-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
				"--client-ca-file=/etc/kubernetes/test-pki/ca.pem",
				"--tls-cert-file=/etc/kubernetes/test-pki/apiserver.pem",
				"--tls-private-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
				"--secure-port=443",
				"--allow-privileged",
				COMPONENT_LOGLEVEL,
			},
			LivenessProbe: componentProbe(8080, "/healthz"),
			Ports: []api.ContainerPort{
				{Name: "https", ContainerPort: 443, HostPort: 443},
				{Name: "local", ContainerPort: 8080, HostPort: 8080},
			},
			Resources: componentResources("250m"),
		}),
		"kube-controller-manager": componentPod(api.Container{
			Name:  "kube-controller-manager",
			Image: HYPERKUBE_IMAGE,
			Command: []string{
				"/hyperkube",
				"controller-manager",
				MASTER,
				CLUSTER_NAME,
				"--root-ca-file=/etc/kubernetes/test-pki/ca.pem",
				"--service-account-private-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
				"--cluster-signing-cert-file=/etc/kubernetes/test-pki/ca.pem",
				"--cluster-signing-key-file=/etc/kubernetes/test-pki/ca-key.pem",
				"--insecure-approve-all-csrs=true",
				COMPONENT_LOGLEVEL,
			},
			LivenessProbe: componentProbe(10252, "/healthz"),
			Resources:     componentResources("200m"),
		}),
		"kube-scheduler": componentPod(api.Container{
			Name:  "kube-controller-manager",
			Image: HYPERKUBE_IMAGE,
			Command: []string{
				"/hyperkube",
				"scheduler",
				MASTER,
				COMPONENT_LOGLEVEL,
			},
			LivenessProbe: componentProbe(10253, "/healthz"),
			Resources:     componentResources("100m"),
		}),
	}

	manifestsPath := path.Join(params.prefixDir, "manifests")
	if err := os.MkdirAll(manifestsPath, 0700); err != nil {
		return err
	}
	for name, spec := range staticPodSpecs {
		serialized, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return err
		}
		if err := util.DumpReaderToFile(bytes.NewReader(serialized), path.Join(manifestsPath, name+".json")); err != nil {
			return err
		}
	}
	return nil
}

func componentResources(cpu string) api.ResourceRequirements {
	return api.ResourceRequirements{
		Requests: api.ResourceList{
			api.ResourceName(api.ResourceCPU): resource.MustParse(cpu),
		},
	}
}

func componentProbe(port int, path string) *api.Probe {
	return &api.Probe{
		Handler: api.Handler{
			HTTPGet: &api.HTTPGetAction{
				Host: "127.0.0.1",
				Path: path,
				Port: intstr.FromInt(port),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
	}
}

func componentPod(container api.Container) api.Pod {
	return api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name:      container.Name,
			Namespace: "kube-system",
			Labels:    map[string]string{"component": container.Name, "tier": "control-plane"},
		},
		Spec: api.PodSpec{
			Containers:      []api.Container{container},
			SecurityContext: &api.PodSecurityContext{HostNetwork: true},
		},
	}
}
