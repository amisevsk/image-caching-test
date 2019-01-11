## Makefile
```bash
# First, set minishift env:
eval $(minishift oc-env)
eval $(minishift docker-env)
# Build Go binary:
make build
# Make docker image:
make docker
# Rollout on openshift:
make rollout
# The three above:
make
# Clean:
make clean
# (Untested) configure minishift cluster:
make configure
```

## Building
```bash
GOOS=linux go build -o ./app . && docker build -t image-caching-test:dev .
```

## Deploying (minishift)

0. Set up some env vars for convenience:
   ```
   CLUSTERROLE_NAME='create-daemonset-cluster'
   CLUSTERROLEBINDING_NAME='daemonset-binding'
   NAMESPACE='daemonset-test'
   ```
1. Switch to desired namespace as admin user:
   ```
   oc project ${NAMESPACE}
   ```
1. Create `clusterrole` for the pod to use:
   ```
   oc create clusterrole ${CLUSTERROLE_NAME} --verb=create,delete,watch, get --resource=daemonset.apps
   ```
1. Create `clusterrolebinding` for the service account:
   ```
   oc create clusterrolebinding ${CLUSTERROLEBINDING_NAME} --clusterrole=${CLUSTERROLE_NAME} --serviceaccount=daemonset-test:daemonset-sa
   ```
1. Create `configmap`:
   ```
   oc create -f configmap.yaml
   ```
1. Create pod and service account:
   ```
   oc process -f app.yaml | oc apply -f -
   ```