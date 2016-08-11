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

package kubelet

import (
	"crypto/x509"
	"io/ioutil"
)

type DiscoveryBase struct {
	ApiVersion string // 'v1alpha1'
	Role       string // 'master' or 'node'
}

type OutOfBandDiscovery struct {
	DiscoveryBase
	Discovery struct {
		ApiServerURLs []string
		CaCertFile    string
	}
}

func (o *OutOfBandDiscovery) Start() {
	// Out of band discovery doesn't need any long-running processes, so this
	// is a no-op.
}

func (o *OutOfBandDiscovery) Discover() ([]string, *x509.Certificate, error) {
	asn1Data, err := ioutil.ReadFile(o.CaCertFile)
	if err != nil {
		return []string{}, nil, err
	}

	caCert, err := x509.ParseCertificate(asn1Data)
	if err != nil {
		return []string{}, nil, err
	}
	return o.ApiServerURLs, caCert, nil
}

type Discovery interface {
	Start()
	Discover() (
		apiServerUrls []string,
		caCert x509.Certificate,
		err error,
	)
}

// TODO implement Discovery methods on GossipDiscovery
type GossipDiscovery struct {
	DiscoveryBase
	Discovery struct {
		Token string
		Peers []string
	}
}
