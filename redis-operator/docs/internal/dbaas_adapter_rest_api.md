With version `1.0.0` of Redis Operator, there are changes in the `Create database` API, specifically in the `settings` section.

The following is the list of predefined keys in the `settings` section for a `Create database` request:

* The `redisDbSettings` is a key-value Redis Configuration map whose values replace the default ones. For more information, refer to [redis.conf](https://download.redis.io/redis-stable/redis.conf). This parameter is optional.

* The `redisDbResources` specifies the resources for new Redis Database that is merged with the ones that are set in the `redis.resources` during installation. For more information, refer to [Redis Parameters](./installation_guide.md#redis-parameters) in the _Redis Installation Procedure_:

  The `redisDbResources.requests.cpu` parameter specifies the minimum number of CPUs the Redis Database should use. This parameter is optional.

  The `redisDbResources.requests.memory` parameter specifies the minimum amount of memory the Redis Database should use. This parameter is optional.

  The `redisDbResources.limits.cpu` parameter the maximum number of CPUs the Redis Database can use. This parameter is optional.

  The `redisDbResources.limits.memory` parameter specifies the maximum amount of memory the Redis Database can use. This parameter is optional.

* The `redisDbNodeSelector` parameter specifies the additional node labels for Redis replica that is merged with the ones that are set in the `redis.nodeLabels` during installation. For more information, refer to [Redis Parameters](./installation_guide.md#redis-parameters) in the _Redis Installation Procedure_. This parameter is optional.

* The `redisDbWaitStartServiceSecond` parameter specifies the duration in seconds during which the Redis adapter tries to connect to the logical database. This parameter is optional. The default value is set to `120`.

# Examples

Run REST request to Adapter Service or create a route on 8080 port.

The default credentials are `dbaas-aggregator/dbaas-aggregator`

* Create database:

  POST /api/v1/dbaas/adapter/redis/databases  
  Auth: -H "Authorization: Basic $(printf "${ADAPTER_USER}:${ADAPTER_PASSWORD}" |base64 )"  
  body: 

    ```
        {
        	"namePrefix": "pref",
            "dbName": "redisdb",
            "password": "predefinedPass"
            "metadata": {
        		"classifier": {
        			"microserviceName": "Service-test",
        			"isServiceDb": true
        		}
        	},
            "settings": {
                "redisDbSettings": {
                    "maxmemory": "123mb"
                },
                "redisDbResources": {
                    "requests": {
                        "cpu": "50m"
                    }, 
                    "limits": {
                        "memory": "222Mi"
                    }
                },
                "redisDbNodeSelector": {
                    "nodeKey": "nodeVal"
                },
                "redisDbWaitStartServiceSecond": 150
            }
        }
    ```

* Get databases list:

   GET /api/v1/dbaas/adapter/redis/databases  
   Auth: -H "Authorization: Basic $(printf "${ADAPTER_USER}:${ADAPTER_PASSWORD}" |base64 )"  


* Drop database:

  POST /api/v1/dbaas/adapter/redis/resources/bulk-drop  
  Auth: -H "Authorization: Basic $(printf "${ADAPTER_USER}:${ADAPTER_PASSWORD}" |base64 )"     
  body: 

  ```
      [
          {
              "kind":"Secret",
              "name":"pref-redisdb-credentials"
          },
          {
              "kind":"ConfigMap",
              "name":"pref-redisdb"
          },
          {
              "kind":"Deployment",
              "name":"pref-redisdb"
          },
          {
              "kind":"Service",
              "name":"pref-redisdb"
          }
      ]
  ```