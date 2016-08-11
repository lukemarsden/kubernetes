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

// This file defines state transition functions for kubelet.
// See: https://github.com/lukemarsden/kubernetes/blob/7e9fe3d4a2d6f3cf4739090b9ffeeb00d85cf365/docs/proposals/cluster-bootstrap-with-gossip.md#kubelet-state-machine

type kubeletInfo struct {
}

// In the style of Rob Pike
type stateFn func(*kubeletInfo) stateFn

func (k *kubeletInfo) run() {
	for state := pendingState; state != nil; {
		state = state(k)
	}
}

// I am a kubelet just born into the world, blinking in the light, unknowing
// even of my role in life, whether to be a node or a master, nor who to trust.
func pendingState(k *kubeletInfo) stateFn {
	return pendingState
}

// I am furnished with a hint as to my role, a key, and perhaps, a friend to
// talk to, especially if I was not made a master.  Try to use the key to find
// out enough information to start a TLS bootstrap.
func gossipingState(k *kubeletInfo) stateFn {
	return pendingState
}

// I am performing TLS bootstrap, growing ever stronger and more confident as
// I locate a trustworthy API server and begin a fine correspondence forthwith.
func bootstrappingState(k *kubeletInfo) stateFn {
	return pendingState
}

// Boom! Whoosh! I am running!
func runningState(k *kubeletInfo) stateFn {
	return pendingState
}
