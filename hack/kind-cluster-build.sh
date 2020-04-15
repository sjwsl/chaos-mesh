#!/usr/bin/env bash
set -e

ROOT=$(unset CDPATH && cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
cd $ROOT

usage() {
    cat <<EOF
This script use kind to create Kubernetes cluster,about kind please refer: https://kind.sigs.k8s.io/
Before run this script,please ensure that:
* have installed docker
* have installed helm
Options:
       -h,--help               prints the usage message
       -n,--name               name of the Kubernetes cluster,default value: kind
       -c,--nodeNum            the count of the cluster nodes,default value: 3
       -k,--k8sVersion         version of the Kubernetes cluster,default value: v1.15.6
       -v,--volumeNum          the volumes number of each kubernetes node,default value: 5
Usage:
    $0 --name testCluster --nodeNum 4 --k8sVersion v1.15.6
EOF
}

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -n|--name)
    clusterName="$2"
    shift
    shift
    ;;
    -c|--nodeNum)
    nodeNum="$2"
    shift
    shift
    ;;
    -k|--k8sVersion)
    k8sVersion="$2"
    shift
    shift
    ;;
    -v|--volumeNum)
    volumeNum="$2"
    shift
    shift
    ;;
    -h|--help)
    usage
    exit 0
    ;;
    *)
    echo "unknown option: $key"
    usage
    exit 1
    ;;
esac
done

clusterName=${clusterName:-kind}
nodeNum=${nodeNum:-3}
k8sVersion=${k8sVersion:-v1.15.6}
volumeNum=${volumeNum:-5}

echo "clusterName: ${clusterName}"
echo "nodeNum: ${nodeNum}"
echo "k8sVersion: ${k8sVersion}"
echo "volumeNum: ${volumeNum}"

source "${ROOT}/hack/lib.sh"

echo "ensuring kind"
hack::ensure_kind
echo "ensuring kubectl"
hack::ensure_kubectl

OUTPUT_BIN=${ROOT}/output/bin
KUBECTL_BIN=${OUTPUT_BIN}/kubectl
HELM_BIN=${OUTPUT_BIN}/helm
KIND_BIN=${OUTPUT_BIN}/kind

echo "############# start create cluster:[${clusterName}] #############"
workDir=${HOME}/kind/${clusterName}
kubeconfigPath=${workDir}/config
mkdir -p ${workDir}

data_dir=${workDir}/data

echo "clean data dir: ${data_dir}"
if [ -d ${data_dir} ]; then
    rm -rf ${data_dir}
fi

configFile=${workDir}/kind-config.yaml

cat <<EOF > ${configFile}
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1alpha3
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServerExtraArgs:
    enable-admission-plugins: NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 5000
    hostPort: 5000
    listenAddress: 127.0.0.1
    protocol: TCP
EOF

for ((i=0;i<${nodeNum};i++))
do
    mkdir -p ${data_dir}/worker${i}
    cat <<EOF >>  ${configFile}
- role: worker
  extraMounts:
EOF
    for ((k=1;k<=${volumeNum};k++))
    do
        mkdir -p ${data_dir}/worker${i}/vol${k}
        cat <<EOF >> ${configFile}
  - containerPath: /mnt/disks/vol${k}
    hostPath: ${data_dir}/worker${i}/vol${k}
EOF
    done
done

echo "start to create k8s cluster"
${KIND_BIN} create cluster --config ${configFile} --image kindest/node:${k8sVersion} --name=${clusterName}
${KIND_BIN} get kubeconfig --name=${clusterName} > ${kubeconfigPath}
export KUBECONFIG=${kubeconfigPath}

${KUBECTL_BIN} apply -f ${ROOT}/manifests/local-volume-provisioner.yaml
${KUBECTL_BIN} apply -f ${ROOT}/manifests/tiller-rbac.yaml

$KUBECTL_BIN create ns chaos-testing

if [[ $(helm version --client --short) == "Client: v2"* ]]; then helm init --service-account=tiller --wait; fi

echo "############# success create cluster:[${clusterName}] #############"

echo "To start using your cluster, run:"
echo "    export KUBECONFIG=${kubeconfigPath}"
echo ""
echo <<EOF
NOTE: In kind, nodes run docker network and cannot access host network.
If you configured local HTTP proxy in your docker, images may cannot be pulled
because http proxy is inaccessible.
If you cannot remove http proxy settings, you can either whitelist image
domains in NO_PROXY environment or use 'docker pull <image> && kind load
docker-image <image>' command to load images into nodes.
EOF
