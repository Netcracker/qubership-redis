## Common Information

Cluster Replicator is a service created to replicate Kubernetes resources between two clusters in a Disaster Recovery (DR) scheme.  
It provides the ability to configure which resources should be replicated by defining them in a replication configuration file. In addition, Cluster Replicator manages micro-services which should be activated, passivated, or scaled depending on the cluster state (Active, Standby, Disable).

In Cloud Core environments, Cluster Replicator is supported out-of-the-box as part of DR functionality (with `DR_MANAGEABLE=true`). For non–Cloud Core applications, replication configuration must be delivered together with the application.

Two DR profiles are supported:
- **Hot Standby**: micro-services are scaled to 1 by default, faster switch, higher reliability.
- **Cold Standby**: micro-services are scaled to 0 by default, resource-efficient, slower switch.
---

### **Prerequisites

Before performing switchover, ensure that:

1. Cluster Replicator is deployed in both Active and Standby clusters.
2. ServiceAccount cluster-replicator has the necessary Role/RoleBinding in target namespaces.
3. Network connectivity between clusters is established (pod networks are reachable).
4. Both clusters have synchronized replication rules and filters.

### **Switchover 

The switchover process on Cluster Replicator is a **planned operational mode change** for maintenance purposes.  
Unlike **failover**, switchover is performed when both sites are operational and healthy.
- **Active**: cluster serves user requests, all services operate normally.  
- **Standby**: cluster suspends requests and internal processes, resources may be scaled down.  

### Key Characteristics

- Planned operational mode change (maintenance scenarios)  
- Both sites remain operational during the process  
- Decision is taken to migrate workload from one site to another  
- No active disruption to applications  
- Data integrity is maintained throughout the process  

### Execution Process

#### Using Site Manager Integration
When `SM_INTEGRATION` is enabled, switchover can be orchestrated through site-manager:  
1. **Access Operational Portal**: Navigate to the `disaster_recovery` project.  
2. **Select Direction**:  
   - `switchover-to-left` — makes left side the new Main site.  
   - `switchover-to-right` — makes right side the new Main site.  
3. **Monitor Pipeline**: Ensure successful completion.  

#### Manual Switchover Process
For manual execution, the switchover follows this sequence:  
1. **Pre-Switchover Status Check**: Verify cluster status:  
   - Main site: `active / done / up` for all services.  
   - Standby site: `standby / done / up` for all services.  
2. **Execute Switchover**: Perform switchover using Cluster Replicator REST API.  
3. **Post-Switchover Verification**: Confirm new state:  
   - Former Main site: `standby / done / up`.  
   - New Main site: `active / done / up`.  

	Important! Data does not replicate between two clusters!

---

## REST API

Cluster Replicator exposes REST API endpoints. Authentication is performed using Kubernetes ServiceAccount tokens (TokenReview).

Before testing API endpoints, ensure that the **service** is accessible inside the **cluster**. Use the internal service address:  

```
http://cluster-replicator.<namespace>.svc.cluster.local:8080
```

---
### Health check

**Request**:
```bash
curl -s http://cluster-replicator.<namespace>.svc.cluster.local:8080/health
```

**Response**:
```json
{"status":"up"}
```

---
### Get current mode
**Request**:

```bash
curl http://cluster-replicator.<namespace>.svc.cluster.local:8080/api/v1/replicator/mode \
  -H "Authorization: Bearer <token>"
```

**Response**:
```json
{"mode":"active","status":"done","message":""}
```

---

### Change mode → Active

**Request**:
```bash
curl -X POST http://cluster-replicator.<namespace>.svc.cluster.local:8080/api/v1/replicator/mode \
  -H "Authorization: Bearer <token>" \
  -d '{"mode":"active"}'
```

**Response**:
```json
{"mode":"active","status":"done","message":""}
```

---

### Change mode → Standby

**Request**:
```bash
curl -X POST http://cluster-replicator.<namespace>.svc.cluster.local:8080/api/v1/replicator/mode \
  -H "Authorization: Bearer <token>" \
  -d '{"mode":"standby"}'
```

**Response**:
```json
{"mode":"standby","status":"done","message":""}
```

---

## Troubleshooting

If a switchover or other DR operation fails, the procedure can be retried.
The retry must be performed in sequence: **active → standby**, then **standby → active**.
Retries are available through the Cluster Replicator REST API.

