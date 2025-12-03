# Korche Placer Service

## Overview

Korche Placer is an **on-demand placement service** designed for Cloud / Edge industrial environments.  
It determines **where to deploy a workload (job/pod)** across a cluster of heterogeneous nodes, considering:

- resource availability (CPU, Memory, Storage),
- energy consumption,
- real-time capabilities,
- node assurance (reliability),
- user-defined scoring weights and hyperparameters.

Unlike a full orchestrator, **Korche Placer does not monitor the cluster**, nor deploy workloads itself.  
It is a **stateless service** invoked **only when a new workload must be placed**.

---

## 🛠️ Running the Service (Docker)

### 1. Download or build the container
```sh
docker pull <your-registry>/korche-placer:latest
