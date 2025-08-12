#Redis monitoring agent    

Works only for DBAAS scheme.

DBAAS creates every db with its own service.

Once another Redis db is created, monitoring pod's config map (telegraf.conf) is supplemented by new db service credentials.

