BINARY_NAME=image-caching-test
DOCKERIMAGE_NAME=image-caching-test
DOCKERIMAGE_TAG=dev
DEPLOYMENT_NAME=che-image-caching

# Configuring openshift
ROLE_NAME='create-daemonset-cluster'
ROLEBINDING_NAME='daemonset-binding'
NAMESPACE='daemonset-test'


all: build docker rollout

build:
	GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go

docker:
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} .

rollout:
	oc rollout latest ${DEPLOYMENT_NAME}

configure:
	oc login -u system:admin
	oc new-project ${NAMESPACE}
	oc project ${NAMESPACE}
	oc create role ${ROLE_NAME} --verb=create,delete,watch,get --resource=daemonset.apps
	oc create rolebinding ${ROLEBINDING_NAME} --role=${ROLE_NAME} --serviceaccount=daemonset-test:che-imagecaching-sa
	oc create -f ./openshift/configmap.yaml
	oc process -f ./openshift/app.yaml | oc apply -f -

clean:
	rm -rf ./bin