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
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	_ "encoding/pem"
	_ "fmt"
	"net"
	"os"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/kubeadm/tlsutil"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// kubeadm is responsible for writing the following file, which kubelet should
// be waiting for. Help user avoid foot-shooting by refusing to write a file
// that has already been written (the kubelet will be up and running in that
// case - they'd need to stop the kubelet, remove the file, and start it again
// in that case).

const (
	KUBELET_BOOTSTRAP_FILE   = "/etc/kubernetes/kubelet-bootstrap.json"
	COMPONENT_LOGLEVEL       = "--v=4"
	SERVICE_CLUSTER_IP_RANGE = "--service-cluster-ip-range=10.16.0.0/12"
	CLUSTER_NAME             = "--cluster-name=kubernetes"
	MASTER                   = "--master=127.0.0.1:8080"
	HYPERKUBE_IMAGE          = "errordeveloper/hyperquick:master"
	MANIFESTS_PATH           = "./manifests/" // TODO use a slice and join it
	PKI_PATH                 = "./pki/"       // TODO use a slice and join it
)

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

func writeStaticPodsOnMaster() error { // TODO it's quite implicitly on master, and it'd be in a separate file, so rename it
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
				"--service-account-key-file=/etc/kubernetes/test-pki/sa-key.pem",
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

	if err := os.MkdirAll(MANIFESTS_PATH, 0700); err != nil {
		return err
	}
	for name, spec := range staticPodSpecs {
		serialized, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return err
		}
		if err := util.DumpReaderToFile(bytes.NewReader(serialized), MANIFESTS_PATH+name+".json"); err != nil {
			return err
		}
	}
	return nil
}

func newCertificateAuthority() (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := tlsutil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	config := tlsutil.CertConfig{
		CommonName: "kubernetes",
	}

	cert, err := tlsutil.NewSelfSignedCACertificate(config, key)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, err
}

func newAPIKeyAndCert(caCert *x509.Certificate, caKey *rsa.PrivateKey, altNames tlsutil.AltNames) (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := tlsutil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	altNames.IPs = append(altNames.IPs, net.ParseIP("10.3.0.1"))
	altNames.DNSNames = append(altNames.DNSNames, []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
	}...)

	config := tlsutil.CertConfig{
		CommonName: "kube-apiserver",
		AltNames:   altNames,
	}
	cert, err := tlsutil.NewSignedCertificate(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, err
	}
	return key, cert, err
}

func newAdminKeyAndCert(caCert *x509.Certificate, caKey *rsa.PrivateKey) (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := tlsutil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	config := tlsutil.CertConfig{
		CommonName: "kubernetes-admin",
	}
	cert, err := tlsutil.NewSignedCertificate(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, err
	}
	return key, cert, err
}

func writeKeysAndCert(name string, key *rsa.PrivateKey, cert *x509.Certificate) error {

	if key != nil {
		if err := util.DumpReaderToFile(bytes.NewReader(tlsutil.EncodePrivateKeyPEM(key)), MANIFESTS_PATH+name+"-key.pem"); err != nil {
			return err
		}
		if pubKey, err := tlsutil.EncodePublicKeyPEM(&key.PublicKey); err == nil {
			if err := util.DumpReaderToFile(bytes.NewReader(pubKey), MANIFESTS_PATH+name+"-pub.pem"); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if cert != nil {
		if err := util.DumpReaderToFile(bytes.NewReader(tlsutil.EncodeCertificatePEM(cert)), MANIFESTS_PATH+name+".pem"); err != nil {
			return err
		}
	}

	return nil
}

func newServiceAccountKey() (*rsa.PrivateKey, error) {
	key, err := tlsutil.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	return key, err
}

func writetPKIAssets() error {
	var (
		err      error
		altNames tlsutil.AltNames // TODO actual SANs
	)

	if err := os.MkdirAll(PKI_PATH, 0700); err != nil {
		return err
	}

	caKey, caCert, err := newCertificateAuthority()
	if err != nil {
		return err
	}

	if err := writeKeysAndCert("ca", caKey, caCert); err != nil {
		return err
	}

	apiKey, apiCert, err := newAPIKeyAndCert(caCert, caKey, altNames)
	if err != nil {
		return err
	}

	if err := writeKeysAndCert("apiserver", apiKey, apiCert); err != nil {
		return err
	}

	saKey, err := newServiceAccountKey()
	if err != nil {
		return err
	}

	if err := writeKeysAndCert("sa", saKey, nil); err != nil {
		return err
	}

	admKey, admCert, err := newAdminKeyAndCert(caCert, caKey)
	if err != nil {
		return err
	}

	if err := writeKeysAndCert("admin", admKey, admCert); err != nil {
		return err
	}

	return err
}
