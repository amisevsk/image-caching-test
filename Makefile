BINARY_NAME=image-caching-test
DOCKERIMAGE_NAME=image-caching-test
DOCKERIMAGE_TAG=dev
DEPLOYMENT_NAME=daemonset-dc

# Configuring openshift
CLUSTERROLE_NAME='create-daemonset-cluster'
CLUSTERROLEBINDING_NAME='daemonset-binding'
NAMESPACE='daemonset-test'



all: build docker rollout

build:
	GOOS=linux go build -o ./bin/${BINARY_NAME} .

docker:
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} .

rollout:
	oc rollout latest ${DEPLOYMENT_NAME}

# Untested
configure:
	oc login -u system:admin
	oc new-project ${NAMESPACE}
	oc project ${NAMESPACE}
	oc create clusterrole ${CLUSTERROLE_NAME} --verb=create,delete,watch,get --resource=daemonset.apps
	oc create clusterrolebinding ${CLUSTERROLEBINDING_NAME} --clusterrole=${CLUSTERROLE_NAME} --serviceaccount=daemonset-test:daemonset-sa
	oc create -f configmap.yaml
	oc process -f app.yaml | oc apply -f -

clean:
	rm -rf ./bin