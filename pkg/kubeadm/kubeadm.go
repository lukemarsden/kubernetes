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
	"os"

	"k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/util/template"
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
	staticPodSpecs = map[string]api.Pod{
		"etcd": api.Pod{
			// TODO this needs a volume
			ApiVersion: "v1",
			Metadata: api.ObjectMeta{
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
						api.LivenessProbe{
							Handler: api.Handler{
								HTTPGet: &api.HTTPGetAction{
									Host: "127.0.0.1",
									Path: "/health",
									Port: 2379,
								},
							},
							InitialDelaySeconds: 15,
							TimeoutSeconds:      15,
						},
						Name: "etcd-container",
						/*
						   "ports": [
						      {
						         "containerPort": 2380,
						         "hostPort": 2380,
						         "name": "serverport"
						      },
						      {
						         "containerPort": 2379,
						         "hostPort": 2379,
						         "name": "clientport"
						      }
						   ],
						   "resources": {
						      "requests": {
						         "cpu": "200m"
						      }
						   }
						*/
					},
				},
				HostNetwork: true,
			},
		},
		"kube-api-server":         &api.Pod{}, // TODO bind-mount certs in
		"kube-controller-manager": &api.Pod{},
		"kube-scheduler":          &api.Pod{},
	}
}

// TODO https://github.com/coreos/bootkube/blob/master/pkg/tlsutil/tlsutil.go
