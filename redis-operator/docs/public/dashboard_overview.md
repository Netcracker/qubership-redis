This section provides information about the dashboards for the Redis servers.

The dashboard for Redis servers is shown in the following image:
![Overview](/docs/public/images/redis_overview.JPG)

**Metrics**

This section describes metrics and their meanings.

You can select specific instance or all instances in the 'Instance' drop-down menu.

**Overview**

* `Redis instances status: Instance` - Displays the of Redis pods.
* `Number of replicas: Instance` - Displays the number of working nodes.
* `Clients: Instance` - Displays the number of clients connected to the selected Redis instance.
* `Memory Usage: Instance` - Displays the RAM usage of the selected Redis instance.
* `CPU Usage: Instance` - Displays CPU usage of the selected Redis instance.
* `Latency: Instance` - Displays the latency (response time) of the selected Redis instance.

**DB Usage**

* `Command Calls: Instance` - Total number of commands processed by the Redis instance over 1m interval.
* `Commands Executed per second: Instance` - Displays the number of operations per second for the selected Redis instance.
* `Total Items per DB: Instance` - Displays totatl amount of items in redis database.
* `Expired/Evicted Keys: Instance` - Displays the nummber of expired and evicted keys on selected redis instance.
* `Expiring vs Not-Expiring Keys: Instance` - Displays the nummber of expiring and not-expiring keys on selected redis instance.
* `Hits / Misses rate: Instance` - The hit rate is the number of cache hits divided by the total number of memory requests over a given time interval. The miss rate is similar that is the total cache misses divided by the total number of memory requests expressed as a percentage over a time interval. If the cache hit ratio is lower than ~0.8 then a significant amount of the requested keys are evicted, expired, or do not exist at all. It is crucial to watch this metric while using Redis as a cache. 
* `CRUD operations: $instance` - Shows status of SET, GET and DEL operations for selected redis instance.

**CPU/RAM Usage**

* `CPU Usage: Instance` - Summary of CPU usage by container Instance. The red line shows available memory volume for the container. The orange line shows requested memory for the container.
* `Memory Usage: Instance` - Summary of Memory usage by container Instance. The red line shows available memory volume for the container. The orange line shows requested memory for the container.
* `Redis Memory Usage: Instance` - Shows the RAM and mapped memory usage of the selected Redis instance.

**Network**
* `Receive/Transmit Bandwidth` - Shows network traffic in bytes per second for the pod.
* `Rate of Received/Transmitted Packets` - Shows network packets for pod.
* `Rate of Received/Transmitted Packets` Dropped - Shows dropped packets for pod.
* `Network I/O: Instance` - Displays the network input and output usage of selected Redis instance. The green line shows the tx_rate for the container, meaning download volumes. The yellow line shows the rx_rate for the container, meaning upload volumes.

