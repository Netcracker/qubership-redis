## To generate CRD:

* Install operator-sdk:

```
RELEASE_VERSION=v0.13.0
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu && sudo mkdir -p /usr/local/bin/ && sudo cp operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk && rm operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
```

* uncomment **replace github.com/Netcracker/qubership-nosqldb-operator-core => ../nosqldb-operator-core**
* run operator-sdk generate k8s
* run operator-sdk generate crds
