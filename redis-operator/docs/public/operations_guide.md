This section describes different commands to work with Redis.

# Table of Contents

  * [Change Password in Redis](#change-password-in-redis)
    * [Change Password for Single Redis Installation](#change-password-for-single-redis-installation)
    * [Change Password for DBaaS Installation](#change-password-for-dbaas-installation)

# Change Password in Redis

This section provides information on how to change the password in Redis.

## Change Password for Single Redis Installation

The password change for single Redis installation can be done in two ways:

* Using the update procedure with the `redis.password` parameter changed. For more information, see [Redis Installation Procedure](/docs/public/installation_guide.md).

    **Note**: In this case, all data is removed from the Redis DB because the pod is restarted. So it is only applicable if you are fine with an empty DB.

* Manual password changing of the running Redis DB pod.

To change the the Redis password, implement the following steps:

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **Redis** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **redis-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with a new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **Redis** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **redis-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Update password in the Redis Pod**

1. Navigate to **Applications > Pods** for Openshift and **Workloads > Pods** for Kubernetes.
1. Navigate to the terminal of the Redis pod.
1. Execute the following command:
   
   ```
   redis-cli -a <current_password> config set requirepass <new_password>
   ```

1. Close the terminal.

## Change Password for DBaaS Installation

Password changing for Redis DB created through DBaaS API can be done in the same way as manual password changing for [Single Redis](#change-password-for-single-redis-installation):

**Update Secret**

For OpenShift:

1. Open the OpenShift UI Console.
1. Select the **Redis** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **<redis_database_name>-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with a new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open the Kubernetes UI Console.
1. Select the **Redis** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **<redis_database_name>-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Update password in the Redis Pod**

1. Navigate to **Applications > Pods** for Openshift and **Workloads > Pods** for Kubernetes.
1. Navigate to the terminal of the **<redis_database_name>** pod.
1. Execute the following command:
   
   ```
   redis-cli -a <current_password> config set requirepass <new_password>
   ```

1. Close the terminal.