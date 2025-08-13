[[_TOC_]]

# Redis Operator

## Repository structure

* `./api` - directory with API for Redis Operator, it contains all parameters definition required for Redis operator to start reconciliation process.
* `./bin` - directory with `controller-gen` binary that is called from Makefile to generate CRD and deep copy methods.
* `./build` - the directory with entrypoint for Docker image of Redis Operator.
* `./charts` - the directory with HELM chart for Redis Operator.
* `./docs` - the directory with actual documentation for the service.
* `./examples` - the directory with deploy parameters examples.
* `./extras` - the directory with files that might be useful for development.
* `./hack` - the directory with licence file.
* `./redis` - the directory with `description.yaml` and `build.sh` file for promotion process.
* `./config` - the directory contains k8s templates, not used in our project.
* `./controllers` - the directory with operator sdk controller.
* `./dbaas/pkg` - the directory with source code.
* `.gitlab-ci.yml` - the CI/CD pipelines configuration.
* `./build.sh` - the entrypoint for build job, it starts docker image build and copies charts.
* `./description.yaml` - descibes buld sructure of Redis Operator docker image.
* `./Dockerfile` - the Dockerfile for Redis Operator docker image.
* `./go.mod` - the go.mod of the project.
* `./go.sum` - the go.sum of the project.
* `./main.go` - the entrypoint of the Redis Operator.
* `./Makefile` - the Makefile to generate CRD and other code.
* `./module_test.go` - the unit tests of Redis Operator.
* `./PROJECT` - the file is used to track the info used to scaffold the project.
* `./renovate.json` - the configuration for renovate bot.

## Artifacts described

The delivery of this repository contains several artifacts:

* **Redis Operator Docker image**, for instance `ghcr.io/netcracker/qubership-redis-operator:feature_initial` - the image to be deployed to Kubernetes/Openshift, contains all logic of the Operator written on Go. Usually we do not need to provide it separately, but for development purpose it might be convinient to update only the image of Operator deployment in Kubernetes/Openshift to test changes


#### Definition of Done

The changes might be marked as fully done if it accepts the following criteria:

1. The ticket's task done.
2. The solution is deployed to dev environment, where it can be tested.
3. Created merge request has:
    1. "Green" pipeline (linter, build, deploy & test jobs are passed).
    3. The description is **fully** filled.

### Deploy to k8s

#### Pure helm

1. Build operator and integration tests, if you need non-master versions.
2. Prepare kubeconfig on you host machine to work with target cluster.
3. Prepare `sample.yaml` file with deployment parameters, which should contains custom docker images if it is needed.
4. Store `sample.yaml` file in `redis-operator/charts/helm/redis-operator` directory.
5. Go to `/charts/helm/redis-operator` directory.
6. Run the following command if you deploy Redis:

     ```sh
     # Run in redis-operator/charts/helm/redis-operator directory
     helm install redis-operator ./ -f sample.yaml -n <TARGET_NAMESPACE>
     ```

## Evergreen strategy

To keep the component up to date, the following activities should be performed regularly:

* Vulnerabilities fixing.
* Redis upgrade.
* Bug-fixing, improvement and feature implementation.

## Useful links

* [Installation Guide](docs/public/installation_guide.md).
* [Troubleshooting Guide](/docs/public/troubleshooting.md).
* [Architecture Guide](/docs/public/architecture.md).
* [Dashboard Overview](/docs/public/dashboard_overview.md).
* [Configuration](/docs/public/configuration.md).
* [Operations Guide](/docs/public/operations_guide.md).