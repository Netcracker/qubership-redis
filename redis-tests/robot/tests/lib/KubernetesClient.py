import kubernetes
from kubernetes.stream import stream
from kubernetes import client, config


class KubernetesClient(object):

    def __init__(self, namespace):
        self.namespace = namespace
        try:
            config.load_incluster_config()
        except config.ConfigException as e:
            print("Can't load incluster kubernetes config. This script is intended to use inside of kubernetes")

        self._api_client = client.ApiClient()
        self._apps_v1_api = client.AppsV1Api(self._api_client)
        self._core_v1_api = client.CoreV1Api(self._api_client)


    def get_deployment_scale(self, name):
        return self._apps_v1_api.read_namespaced_deployment_scale(
            name=name, namespace=self.namespace)

    def set_deployment_scale(self, name, namespace, scale):
        self._apps_v1_api.replace_namespaced_deployment_scale(name=name,
                                                              namespace=self.namespace,
                                                              body=scale,
                                                              pretty='true')

    def get_deployments(self):
        return self._apps_v1_api.list_namespaced_deployment(namespace=self.namespace)

    def get_deployments_for_service(self, service, label_name='clusterName'):
        """
        This method gets all found (active/inactive) deployments name for given service.
        *Args:*\n
            _namespace_ (str) - OpenShift project name;\n
            _service_ (str) - service name;\n
        *Return:*\n
            list(str) - list of deployments name for given service
        """
        all_deployments_in_project = self.get_deployments()
        deployments = []
        for deployment in all_deployments_in_project.items:
            if deployment.spec.template.metadata.labels.get(label_name, '') == service:
                deployments.append(deployment.metadata.name)
        return deployments

    def get_pods(self):
        return self._core_v1_api.list_namespaced_pod(namespace=self.namespace)

    def delete_pod(self, name):
        body = kubernetes.client.V1DeleteOptions()
        self._core_v1_api.delete_namespaced_pod(name, body)

    def execute_command_in_pod(self, pod_name, command):
        exec_cmd = command.split(" ")
        return stream(self._core_v1_api.connect_get_namespaced_pod_exec,
                      pod_name,
                      self.namespace,
                      command=exec_cmd,
                      stderr=True,
                      stdin=False,
                      stdout=True,
                      tty=False,
                      _preload_content=False)
