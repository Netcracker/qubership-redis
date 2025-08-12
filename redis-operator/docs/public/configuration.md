This section provides information about Service Ports and Dependencies.

# List of Service Ports and Dependencies

Netcracker Redis Service consists Redis, Redis Dbaas Adapter.
The list of service ports and dependencies are described in the following table:

|Component|Exposed Ports|Dependencies (Used Ports)|
|---------|-------------|-------------------------|
|redis|`6379/TCP` The Redis server port, used by client applications to access database.|Redis-Client - `8080/TCP, 8443/TCP (TLS)`|
|redis DBaaS Adapter|`8080/TCP, 8443/TCP (TLS)` The DBaaS API port, used by the DBaaS aggregator to manage redis database.|Redis - `6379/TCP`, DBaaS Aggregator - `8080/TCP, 8443/TCP (TLS)`|


# Rules require the password to have:

* At least 8 characters
* At least one uppercase character
* At least one lowercase character
* At least one number
* At least one special character
* Cannot contain the user's email address or the reverse of the email address.
* Cannot have more than three repeating characters.

For more detailed information, refer to the official Redis security documentation.
