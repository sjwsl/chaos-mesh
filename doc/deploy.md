# Deploy Chaos Mesh

This document describes how to deploy Chaos Mesh for performing chaos experiments 
on your application on Kubernetes. 

## Prerequisites

Before deploying Chaos Mesh, make sure the following items have been installed:

- Kubernetes >= v1.12
- [RBAC](https://kubernetes.io/docs/admin/authorization/rbac) enabled (optional)
- [Helm](https://helm.sh/) version >= v2.8.2

## Step 1: Get Chaos Mesh

```bash
git clone https://github.com/pingcap/chaos-mesh.git
cd chaos-mesh/
```

## Step 2: Create custom resource type

To use Chaos Mesh, you must first create the related custom resource type.

```bash
kubectl apply -f manifests/crd.yaml
```

## Step 3: Install Chaos Mesh

Depending on your environment, there are following methods of installing Chaos Mesh: 

* Install in docker environment

```bash
# create namespace chaos-testing
kubectl create ns chaos-testing
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing
# check Chaos Mesh pods installed
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```

* Install in containerd environment (Kind)

```bash
# create namespace chaos-testing
kubectl create ns chaos-testing
# helm 2.X
helm install helm/chaos-mesh --name=chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
# helm 3.X
helm install chaos-mesh helm/chaos-mesh --namespace=chaos-testing --set chaosDaemon.runtime=containerd --set chaosDaemon.socketPath=/run/containerd/containerd.sock
# check Chaos Mesh pods installed
kubectl get pods --namespace chaos-testing -l app.kubernetes.io/instance=chaos-mesh
```

> **Note:**
>
> Due to current development status of Chaos Dashboard, it is not installed by default. If you want to try it out, add `--set dashboard.create=true` in the helm commands above. Please refer to [Configuration](../helm/chaos-mesh/README.md#parameters) for more information.

After executing the above commands, you should be able to see the output indicating that all Chaos Mesh pods are up and running. Otherwise, please check the current environment according to the prompt message or send us an [issue](https://github.com/pingcap/chaos-mesh/issues) for help.

## Next steps

Refer to [Run Chaos Mesh](run_chaos_mesh.md).

