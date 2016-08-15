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
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// kubeadm is responsible for writing the following file, which kubelet should
// be waiting for. Help user avoid foot-shooting by refusing to write a file
// that has already been written (the kubelet will be up and running in that
// case - they'd need to stop the kubelet, remove the file, and start it again
// in that case).

const KUBELET_BOOTSTRAP_FILE = "/etc/kubernetes/kubelet-bootstrap.json"

func writeParamsIfNotExists(params *BootstrapParams) error {
	serialized, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// Create and open the file, only if it does not already exist.
	f, err := os.OpenFile(
		KUBELET_BOOTSTRAP_FILE,
		os.O_CREATE|os.O_WRONLY|os.O_EXCL,
		0600,
	)
	defer f.Close()

	_, err = f.Write(serialized)
	if err != nil {
		return err
	}
	return nil
}

func writeStaticPodsOnMaster() error {
	staticPodSpecs := map[string]api.Pod{
		"etcd": api.Pod{
			// TODO this needs a volume
			ObjectMeta: api.ObjectMeta{
				Name:      "etcd-server",
				Namespace: "kube-system",
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Command: []string{
							"/bin/sh",
							"-c",
							"/usr/local/bin/etcd --listen-peer-urls http://127.0.0.1:2380 --addr 127.0.0.1:2379 --bind-addr 127.0.0.1:2379 --data-dir /var/etcd/data",
						},
						Image: "gcr.io/google_containers/etcd:2.2.1",
						LivenessProbe: &api.Probe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Host: "127.0.0.1",
									Path: "/health",
									Port: intstr.FromInt(2379),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      15,
						},
						Name: "etcd-container",
						Ports: []api.ContainerPort{
							{Name: "serverport", ContainerPort: 2380, HostPort: 2380},
							{Name: "clientport", ContainerPort: 2379, HostPort: 2379},
						},
						Resources: api.ResourceRequirements{
							Requests: api.ResourceList{
								api.ResourceName(api.ResourceCPU): resource.MustParse("200m"),
							},
						},
					},
				},
				SecurityContext: &api.PodSecurityContext{HostNetwork: true},
			},
		},
		"kube-apiserver": api.Pod{
			// TODO bind-mount certs in
			ObjectMeta: api.ObjectMeta{
				Name:      "kube-apiserver",
				Namespace: "kube-system",
				Labels:    map[string]string{"component": "kube-apiserver", "tier": "control-plane"},
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Command: []string{
							"/hyperkube",
							"apiserver",
							"--address=127.0.0.1",
							"--etcd-servers=http://127.0.0.1:2379",
							"--cloud-provider=fake",
							"--admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,ResourceQuota",
							"--service-cluster-ip-range=10.16.0.0/12",
							"--service-account-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
							"--client-ca-file=/etc/kubernetes/test-pki/ca.pem",
							"--tls-cert-file=/etc/kubernetes/test-pki/apiserver.pem",
							"--tls-private-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
							"--secure-port=443",
							"--allow-privileged",
							"--v=4",
						},
						Image: "errordeveloper/hyperquick:master",
						LivenessProbe: &api.Probe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Host: "127.0.0.1",
									Path: "/healthz",
									Port: intstr.FromInt(8080),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      15,
						},
						Name: "kube-apiserver",
						Ports: []api.ContainerPort{
							{Name: "https", ContainerPort: 443, HostPort: 443},
							{Name: "local", ContainerPort: 8080, HostPort: 8080},
						},
						Resources: api.ResourceRequirements{
							Requests: api.ResourceList{
								api.ResourceName(api.ResourceCPU): resource.MustParse("250m"),
							},
						},
					},
				},
				SecurityContext: &api.PodSecurityContext{HostNetwork: true},
			},
		},
		"kube-controller-manager": api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name:      "kube-controller-manager",
				Namespace: "kube-system",
				Labels:    map[string]string{"component": "kube-controller-manager", "tier": "control-plane"},
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Command: []string{
							"/hyperkube",
							"controller-manager",
							"--master=127.0.0.1:8080",
							"--cluster-name=kubernetes",
							"--root-ca-file=/etc/kubernetes/test-pki/ca.pem",
							"--service-account-private-key-file=/etc/kubernetes/test-pki/apiserver-key.pem",
							"--cluster-signing-cert-file=/etc/kubernetes/test-pki/ca.pem",
							"--cluster-signing-key-file=/etc/kubernetes/test-pki/ca-key.pem",
							"--insecure-approve-all-csrs=true",
							"--v=4",
						},
						Image: "errordeveloper/hyperquick:master",
						LivenessProbe: &api.Probe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Host: "127.0.0.1",
									Path: "/healthz",
									Port: intstr.FromInt(10252),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      15,
						},
						Name: "kube-controller-manager",
						Resources: api.ResourceRequirements{
							Requests: api.ResourceList{
								api.ResourceName(api.ResourceCPU): resource.MustParse("200m"),
							},
						},
					},
				},
				SecurityContext: &api.PodSecurityContext{HostNetwork: true},
			},
		},
		"kube-scheduler": api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name:      "kube-scheduler",
				Namespace: "kube-system",
				Labels:    map[string]string{"component": "kube-scheduler", "tier": "control-plane"},
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Command: []string{
							"/hyperkube",
							"scheduler",
							"--master=127.0.0.1:8080",
							"--v=4",
						},
						Image: "errordeveloper/hyperquick:master",
						LivenessProbe: &api.Probe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Host: "127.0.0.1",
									Path: "/healthz",
									Port: intstr.FromInt(10253),
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      15,
						},
						Name: "kube-controller-manager",
						Resources: api.ResourceRequirements{
							Requests: api.ResourceList{
								api.ResourceName(api.ResourceCPU): resource.MustParse("100m"),
							},
						},
					},
				},
				SecurityContext: &api.PodSecurityContext{HostNetwork: true},
			},
		},
	}
	serialized, err := json.Marshal(staticPodSpecs)
	if err == nil {
		fmt.Printf("staticPodSpecs: %q", serialized)
	}
	return err
}

// TODO https://github.com/coreos/bootkube/blob/master/pkg/tlsutil/tlsutil.go
