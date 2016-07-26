# Proposal: New turnup UX

This proposal aims to capture the goals and proposed plan of action of SIG-cluster-lifecycle, so that it can be discussed textually, in proposal form, as well as in our weekly meetings.

## Motivation

Kubernetes is hard to install, and there are many different ways to do it today.

None of them are excellent.

We believe this is hindering adoption.

## Goals

Have one recommended, official, tested, "happy path" which will enable a majority of new and existing Kubernetes users to:

* Kick the tires and easily turn up a new cluster on infrastructure of their choice
* Get a reasonably secure, production-ready cluster, with reasonable defaults and a range of easily-installable add-ons

We plan to do so by improving and simplifying Kubernetes itself, rather than building lots of tooling which "wraps" Kubernetes by poking all the bits into the right place.

Where possible, Kubernetes should be used to manage itself.

## Scope of project

There are logically 3 steps to deploying a Kubernetes cluster:

1. Provisioning: Getting some servers - these may be VMs on a developer's workstation, VMs in public clouds, or bare-metal servers in a user's data center.
2. Bootstrapping and discovery: Installing the Kubernetes core components on those servers (`etcd`, `kubelet`, etc) - which components get installed can differs between masters and nodes - and bootstrapping the cluster to a state of basic liveness, including allowing each server in the cluster to discover other servers: for example teaching `etcd` servers about their peers, installing SSL certificates, etc.
3. Add-ons: Now that basic cluster functionality is working, installing add-ons such as DNS or a pod network (should be possible using `kubectl apply`).

Notably, this project ("SIG-cluster-lifecycle") is only working on radically improving improving 2 and 3 from the perspective of users typing commands directly into root shells of servers.
The reason for this is that there are a great many different ways of provisioning servers, and users will already have their own preferences.

What's more, once we've improved the user experience of 2 and 3, it will make the job of tools that want to do all 3 much easier.

## Top-down view: UX

The `kubelet` will be taught how to bootstrap a Kubernetes cluster.
This makes sense because `kubelet` already needs to be installed on all your hosts.
So `kubelet` can be the thing that gets installed via downloading a binary, or an OS package.

Self-hosting the `kubelet` is out of scope for now.

This section highlights the developer-facing view of installing Kubernetes via this new CLI.

* Based on Joe Beda's UX, see link below.
* Takes into account:

  * Multiple discovery mechanisms, e.g. a bootstrap network like Kubernetes-Anywhere, or a public discovery service like etcd

* This is a straw man!
  This UX proposal is intended to promote discussion: are these the right knobs?

```
workstation$ kubelet --help

Kubelet: core component of Kubernetes.
I also know how to bootstrap Kubernetes clusters.

Run me on servers to transform them into a Kubernetes cluster suitable either
for tire-kicking or production usage.

This tool assumes you already have some servers running Linux and Docker.

I can deploy masters (where the Kubernetes control plane runs), and nodes
(where your containers get deployed to), or servers that do both at the same
time (useful for smaller clusters). Servers can either be of type "node",
"master", or "master-and-node".

All kubernetes components will be deployed in containers.
An etcd cluster will be created for you.

Global arguments:

--disco   which service discovery mechanism to use for kubernetes bootstrap
          (choose from "weave", "dns", "token", "consul", default: "weave").

--pki     certificate provider, default "auto" to ask the discovery mechanism
          to bootstrap certs for you when you "init" (choose from "vault",
          "amazon-cm", "containers", "token").

--net     which pod network to create (choose from "weave", "flannel", default:
          "weave")

You must use the same global arguments on all nodes in your cluster.

Subcommands:
    init         Run this on the first server you deploy onto.

    join         Run this on other servers to join an existing cluster.

    user-create  Create e.g. 'joe.kubecred' file which can be imported into kubectl


Examples:

Run a single node cluster for development

localhost$ kube init master-and-node
localhost$ kube toolbox
toolbox$ kubectl get nodes
NAME        STATUS    AGE
node        Ready     1m


Using weave net and disco (the defaults) to create a
small 3-node multi-master test cluster

10.0.0.1$ kubelet init master-and-node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.2$ kubelet join master-and-node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.3$ kubelet join master-and-node 10.0.0.1,10.0.0.2,10.0.0.3


10.0.0.1$ kubelet toolbox
toolbox$ kubectl get nodes
NAME        STATUS    AGE
deadbeef    Ready     1m
cafebabe    Ready     1m
beefcafe    Ready     1m


Using weave net and disco (the defaults) to create a
production 9-node test cluster with better resilience
against 10.0.0.1 failing during bootstrapping

# 3 masters
10.0.0.1$ kubelet init master 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.2$ kubelet join master 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.3$ kubelet join master 10.0.0.1,10.0.0.2,10.0.0.3

# 6 nodes
10.0.0.4$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.5$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.6$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.7$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.8$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3
10.0.0.9$ kubelet join node 10.0.0.1,10.0.0.2,10.0.0.3


Using flannel pod network and public discovery service

10.0.0.1$ kubelet --net=flannel --disco=token init master-and-node
Connecting to discovery.k8s.io...
Rendezvous token: deadbeef
10.0.0.2$ kubelet --net=flannel --disco=token join master-and-node deadbeef
10.0.0.3$ kubelet --net=flannel --disco=token join master-and-node deadbeef
```

See also: [Joe Beda's "K8s the hard way easier"](https://docs.google.com/document/d/1lJ26LmCP-I_zMuqs6uloTgAnHPcuT7kOYtQ7XSgYLMA/edit#heading=h.ilgrv18sg5t) which combines Kelsey's "Kubernetes the hard way" with history of proposed UX at the end (scroll all the way down to the bottom).

## Bottom-up view: what needs working on in Kubernetes core

Each of these categories of work may link to sub-proposals as needed.

* PKI bootstrapping - creating and distributing certificates as-needed
* Others, please add!

## Prototype of UX

Ilya and Luke at Weaveworks have *prototyped* the desired user experience listed described in the "Top-down view" section.

This prototype is called `kube-alpha` for lack of a better name.
It is currently living in (Ilya's fork of `kubernetes-anywhere)[https://github.com/errordeveloper/kubernetes-anywhere/tree/kube-alpha/phase2/kube-alpha].

Run `dev-init.sh` and then `dev.sh` on a machine with `docker-machine` installed and you'll get guided through the current prototype UX for installing a 2-node Kubernetes cluster on two Docker Machine VMs.

The purpose of this prototype is to *demonstrate that the desired UX is possible* and to enable discussion around it.
The prototype is written in Golang because the final version will need to be integrated into the `kubelet`.

The techniques used in the prototype are not intended to be final!

Experimenting in this way allows us to start fleshing out the Golang interfaces that will make sense to provide to .

This is intended to provide a period of a month or so where different implementations of these interfaces can be tried and tested out.
So for example, discovery can be provided by a Weave network, by a remote discovery API, or by DNS.

As we implement different implementations of the interfaces, we can change the interfaces as necessary.

Ultimately, this project should aim to decide on a handful of implementations of these interfaces, including some reasonable defaults.
