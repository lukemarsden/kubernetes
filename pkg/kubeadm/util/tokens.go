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
package kubeadmutil

import (
	"encoding/hex"
	"fmt"
	"strings"

	kubeadmapi "k8s.io/kubernetes/pkg/kubeadm/api"
	utilrand "k8s.io/kubernetes/pkg/util/rand"
)

const (
	TokenIDLen = 6
	TokenBytes = 8
)

func GenerateToken(params *kubeadmapi.BootstrapParams) error {
	fmt.Println("Generating token...")
	tokenID, token := utilrand.HexString(TokenIDLen), utilrand.HexString(TokenBytes*2)
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return err
	}

	params.Discovery.TokenID = tokenID
	params.Discovery.BearerToken = token
	params.Discovery.Token = tokenBytes
	params.Discovery.GivenToken = fmt.Sprintf("%s.%s", token, tokenID)
	return nil
}

func UseGivenTokenIfValid(params *kubeadmapi.BootstrapParams) (bool, error) {
	if params.Discovery.GivenToken == "" {
		return false, nil
	}
	givenToken := strings.Split(strings.ToLower(params.Discovery.GivenToken), ".")
	// TODO print desired format
	// TODO could also print more specific messages in each case
	invalidErr := "<util/tokens> provided token is invalid - %s"
	if len(givenToken) != 2 {
		return false, fmt.Errorf(invalidErr, "not in 2-part dot-separated format")
	}
	if len(givenToken[0]) != TokenIDLen {
		return false, fmt.Errorf(invalidErr, fmt.Sprintf(
			"length of first part is incorrect [%d (given) != %d (expected) ]",
			len(givenToken[0]), TokenIDLen))
	}
	tokenBytes, err := hex.DecodeString(givenToken[1])
	if err != nil {
		return false, fmt.Errorf(invalidErr, err)
	}
	if len(tokenBytes) != TokenBytes {
		return false, fmt.Errorf(invalidErr, fmt.Sprintf(
			"length of second part is incorrect [%d (given) != %d (expected)]",
			len(tokenBytes), TokenBytes))
	}
	params.Discovery.TokenID = givenToken[0]
	params.Discovery.BearerToken = givenToken[1]
	params.Discovery.Token = tokenBytes
	return true, nil // given and valid
}
