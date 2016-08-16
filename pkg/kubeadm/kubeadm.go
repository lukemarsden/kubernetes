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

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// kubeadm is responsible for writing the following file, which kubelet should
// be waiting for. Help user avoid foot-shooting by refusing to write a file
// that has already been written (the kubelet will be up and running in that
// case - they'd need to stop the kubelet, remove the file, and start it again
// in that case).

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

// TODO https://github.com/coreos/bootkube/blob/master/pkg/tlsutil/tlsutil.go
