This section provides information about the dashboards for the Redis servers.

The dashboard for Redis servers is shown in the following image:
![Overview](/docs/images/redis_overview.PNG)

**Metrics**

This section describes metrics and their meanings.

**Overview**

* `Redis instances status (1=UP, 0=DOWN)` - Shows the UP/DOWN status of Redis pods.
* `Health` - Shows the UP/DOWN status of the selected Redis instance.
* `Clients` - Shows the number of clients connected to the selected Redis instance.
* `Memory Usage` - Shows the RAM usage of the selected Redis instance.
* `CPU Usage` - Shows CPU usage of the selected Redis instance.
* `Latency` - Shows the latency (response time) of the selected Redis instance.

**DB Usage**

* `Command Calls` - Total number of commands processed by the Redis instance over 1m interval.
* `Commands Executed per second` - Shows the number of operations per second for the selected Redis instance.
* `Network I/O` - Shows the network input and output usage of selected Redis instance. The green line shows the tx_rate for the container, meaning download volumes. The yellow line shows the rx_rate for the container, meaning upload volumes.
* `Hits / Misses rate` - The hit rate is the number of cache hits divided by the total number of memory requests over a given time interval. The miss rate is similar that is the total cache misses divided by the total number of memory requests expressed as a percentage over a time interval. If the cache hit ratio is lower than ~0.8 then a significant amount of the requested keys are evicted, expired, or do not exist at all. It is crucial to watch this metric while using Redis as a cache. 

**CPU/RAM Usage**

**Panel `Memory Usage: Instance` is multiplied by parameter `Instance`.**

* `Memory Usage: Instance` - Summary of Memory usage by container Instance. The red line shows available memory volume for the container. The orange line shows requested memory for the container.
* `Redis Memory Usage` - Shows the RAM and mapped memory usage of the selected Redis instance.