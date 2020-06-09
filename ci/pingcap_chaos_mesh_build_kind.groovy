//
// E2E Jenkins file.
//

import groovy.transform.Field

@Field
def podYAML = '''
apiVersion: v1
kind: Pod
metadata:
  labels:
    # we pretend as tidb-operator in order not to meet tidb-operator-e2e job in the same node
    app: tidb-operator-e2e
spec:
  containers:
  - name: main
    image: gcr.io/k8s-testimages/kubekins-e2e:v20200311-1e25827-master
    command:
    - runner.sh
    # Clean containers on TERM signal in root process to avoid cgroup leaking.
    # https://github.com/pingcap/tidb-operator/issues/1603#issuecomment-582402196
    - exec
    - bash
    - -c
    - |
      function clean() {
        echo "info: clean all containers to avoid cgroup leaking"
        docker kill $(docker ps -q) || true
        docker system prune -af || true
      }
      trap clean TERM
      sleep 1d & wait
    # we need privileged mode in order to do docker in docker
    securityContext:
      privileged: true
    env:
    - name: DOCKER_IN_DOCKER_ENABLED
      value: "true"
    resources:
      requests:
        memory: "4Gi"
        cpu: 4
        ephemeral-storage: "10Gi"
      limits:
        memory: "8Gi"
        cpu: 8
        ephemeral-storage: "50Gi"
    # kind needs /lib/modules and cgroups from the host
    volumeMounts:
    - mountPath: /lib/modules
      name: modules
      readOnly: true
    - mountPath: /sys/fs/cgroup
      name: cgroup
    # dind expects /var/lib/docker to be volume
    - name: docker-root
      mountPath: /var/lib/docker
    # legacy docker path for cr.io/k8s-testimages/kubekins-e2e
    - name: docker-graph
      mountPath: /docker-graph
  volumes:
  - name: modules
    hostPath:
      path: /lib/modules
      type: Directory
  - name: cgroup
    hostPath:
      path: /sys/fs/cgroup
      type: Directory
  - name: docker-root
    emptyDir: {}
  - name: docker-graph
    emptyDir: {}
  affinity:
    # running on nodes for chaos-mesh only
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: ci.pingcap.com
            operator: In
            values:
            # we pretend as tidb-operator in order not to meet tidb-operator-e2e job in the same node
            - tidb-operator
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: app
              operator: In
              values:
              # we pretend as tidb-operator in order not to meet tidb-operator-e2e job in the same node
              - tidb-operator-e2e
          topologyKey: kubernetes.io/hostname
'''

def build(String name, String code) {
	podTemplate(yaml: podYAML) {
		node(POD_LABEL) {
			container('main') {
				def WORKSPACE = pwd()
				def ARTIFACTS = "${WORKSPACE}/go/src/github.com/pingcap/chaos-mesh/_artifacts"
				try {
					dir("${WORKSPACE}/go/src/github.com/pingcap/chaos-mesh") {
						unstash 'chaos-mesh'
						stage("Debug Info") {
							println "debug host: 172.16.5.15"
							println "debug command: kubectl -n jenkins-ci exec -ti ${NODE_NAME} bash"
							sh """
							echo "====== shell env ======"
							echo "pwd: \$(pwd)"
							env
							echo "====== go env ======"
							go env
							echo "====== docker version ======"
							docker version
							"""
						}
						stage('Run') {
							ansiColor('xterm') {
								sh """
								export GOPATH=${WORKSPACE}/go
								export ARTIFACTS=${ARTIFACTS}
								${code}
								"""
							}
						}
					}
				} finally {
                    dir(ARTIFACTS) {
						sh """#!/bin/bash
						echo "info: change ownerships for jenkins"
						chown -R 1000:1000 .
						echo "info: print total size of artifacts"
						du -sh .
						echo "info: list all files"
						find .
						echo "info: moving all artifacts into a sub-directory"
						shopt -s extglob
						mkdir ${name}
						mv !(${name}) ${name}/
						"""
						archiveArtifacts artifacts: "${name}/**", allowEmptyArchive: true
						junit testResults: "${name}/*.xml", allowEmptyResults: true
					}
				}
			}
		}
	}
}

def getChangeLogText() {
	def changeLogText = ""
	for (int i = 0; i < currentBuild.changeSets.size(); i++) {
		for (int j = 0; j < currentBuild.changeSets[i].items.length; j++) {
			def commitId = "${currentBuild.changeSets[i].items[j].commitId}"
			def commitMsg = "${currentBuild.changeSets[i].items[j].msg}"
			changeLogText += "\n" + "`${commitId.take(7)}` ${commitMsg}"
		}
	}
	return changeLogText
}

def call(BUILD_BRANCH, CREDENTIALS_ID) {
	timeout (time: 2, unit: 'HOURS') {

	def GITHASH
	def UCLOUD_OSS_URL = "http://pingcap-dev.hk.ufileos.com"
	def BUILD_URL = "git@github.com:pingcap/chaos-mesh.git"
	def PROJECT_DIR = "go/src/github.com/pingcap/chaos-mesh"

	catchError {
		node('build_go1130_memvolume') {
			container("golang") {
				def WORKSPACE = pwd()
				dir("${PROJECT_DIR}") {
					deleteDir()

					stage('Checkout') {
						checkout changelog: false,
						poll: false,
						scm: [
							$class: 'GitSCM',
							branches: [[name: "${BUILD_BRANCH}"]],
							doGenerateSubmoduleConfigurations: false,
							extensions: [],
							submoduleCfg: [],
							userRemoteConfigs: [[
								credentialsId: "${CREDENTIALS_ID}",
								refspec: '+refs/heads/*:refs/remotes/origin/* +refs/pull/*:refs/remotes/origin/pr/*',
								url: "${BUILD_URL}",
							]]
						]

						GITHASH = sh(returnStdout: true, script: "git rev-parse HEAD").trim()
					}

					stage("Build") {
						ansiColor('xterm') {
							sh """
							make ensure-kubebuilder
							make ensure-kustomize
							make binary
							make manifests
							make e2e-build
							"""
						}
					}

					stage("Prepare for e2e") {
						ansiColor('xterm') {
							sh """
							hack/prepare-e2e.sh
							"""
						}
					}

					stash excludes: "vendor/**,deploy/**", name: "chaos-mesh"
				}
			}
		}

		def GLOBALS = "SKIP_BUILD=y IMAGE_TAG=${GITHASH} GINKGO_NO_COLOR=y"
		def artifacts = "go/src/github.com/pingcap/chaos-mesh/artifacts"
		def builds = [:]
		builds["E2E v1.12.10"] = {
                build("v1.12", "${GLOBALS} GINKGO_NODES=6 KUBE_VERSION=v1.12.10 ./hack/e2e.sh -- --ginkgo.focus='Basic'")
        }
		builds.failFast = false
		parallel builds

		currentBuild.result = "SUCCESS"
	}

	stage('Summary') {
		def CHANGELOG = getChangeLogText()
		def duration = ((System.currentTimeMillis() - currentBuild.startTimeInMillis) / 1000 / 60).setScale(2, BigDecimal.ROUND_HALF_UP)
		def slackmsg = "[#${env.ghprbPullId}: ${env.ghprbPullTitle}]" + "\n" +
		"${env.ghprbPullLink}" + "\n" +
		"${env.ghprbPullDescription}" + "\n" +
		"Integration Common Test Result: `${currentBuild.result}`" + "\n" +
		"Elapsed Time: `${duration} mins` " + "\n" +
		"${CHANGELOG}" + "\n" +
		"${env.RUN_DISPLAY_URL}"

		if (currentBuild.result != "SUCCESS") {
			slackSend channel: '#cloud_jenkins', color: 'danger', teamDomain: 'pingcap', tokenCredentialId: 'slack-pingcap-token', message: "${slackmsg}"
			return
		}

		slackSend channel: '#cloud_jenkins', color: 'good', teamDomain: 'pingcap', tokenCredentialId: 'slack-pingcap-token', message: "${slackmsg}"
	}

    }
}

return this

// vim: noet
