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
	"net"
	"os"
	"path"

	"k8s.io/kubernetes/pkg/kubeadm/tlsutil"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

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

func writeKeysAndCert(pkiPath string, name string, key *rsa.PrivateKey, cert *x509.Certificate) error {

	if key != nil {
		if err := util.DumpReaderToFile(bytes.NewReader(tlsutil.EncodePrivateKeyPEM(key)), path.Join(pkiPath, name+"-key.pem")); err != nil {
			return err
		}
		if pubKey, err := tlsutil.EncodePublicKeyPEM(&key.PublicKey); err == nil {
			if err := util.DumpReaderToFile(bytes.NewReader(pubKey), path.Join(pkiPath, name+"-pub.pem")); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if cert != nil {
		if err := util.DumpReaderToFile(bytes.NewReader(tlsutil.EncodeCertificatePEM(cert)), path.Join(pkiPath, name+".pem")); err != nil {
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

func generateAndWritePKIAssets(params *BootstrapParams) error {
	var (
		err      error
		altNames tlsutil.AltNames // TODO actual SANs
	)

	pkiPath := path.Join(params.prefixDir, "pki")
	if err := os.MkdirAll(pkiPath, 0700); err != nil {
		return err
	}

	caKey, caCert, err := newCertificateAuthority()
	if err != nil {
		return err
	}

	if err := writeKeysAndCert(pkiPath, "ca", caKey, caCert); err != nil {
		return err
	}

	apiKey, apiCert, err := newAPIKeyAndCert(caCert, caKey, altNames)
	if err != nil {
		return err
	}

	if err := writeKeysAndCert(pkiPath, "apiserver", apiKey, apiCert); err != nil {
		return err
	}

	saKey, err := newServiceAccountKey()
	if err != nil {
		return err
	}

	if err := writeKeysAndCert(pkiPath, "sa", saKey, nil); err != nil {
		return err
	}

	admKey, admCert, err := newAdminKeyAndCert(caCert, caKey)
	if err != nil {
		return err
	}

	if err := writeKeysAndCert(pkiPath, "admin", admKey, admCert); err != nil {
		return err
	}

	return err
}
