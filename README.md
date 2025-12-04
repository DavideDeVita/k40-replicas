
# KOrche Placer Service

## Overview

The **KOrche Placer** is an on-demand placement service designed for Cloud/Edge industrial workloads.  
It evaluates the current cluster state and determines which worker nodes should host a new workload (pod/job), taking into account multiple objectives such as:

- Resource availability (CPU, Memory, Storage)
- Criticality and Replicas
- Real-time elegibility
- Energy consumption

This service does not perform orchestration, autoscaling, or real-time metric collection.  
It simply answers the question:

**"Given this cluster and this workload, where should it be placed?"**

Call the service when you need a placement decision. It does nothing unless invoked.

---

## REST API

### Endpoint

```
POST /place
Content-Type: application/json
```

### What it does

You send:

- The description of your cluster and its worker nodes
- The workload (pod/job) you want to deploy
- The desired algorithm and tuning parameters (optional)

The service responds with:

- Whether the workload can be placed
- How many replicas are needed
- Which nodes should host them
- Updated resource usage for each selected node
- A multi-objective score and breakdown of sub-scores

---

## Running the Service via Docker

### Prerequisites

- Docker installed and running
- Access to the `korche-placer` image

You can load the image with:
```powershell
docker load -i "korche-placer.tar"
```


### Run the container

```
docker run -p 8080:8080 korche-placer
```

### Optional: Produce placement output files

To make the service save every placement response as a JSON file (rather than writing it on STDOUT):

Create a directory:

```
mkdir out
```

Mount it into the container:

(in PowerShell)
```powershell
docker run -p 8080:8080 -v ${PWD}\out:/app/out korche-placer
```

(in Bash)
```bash
docker run -p 8080:8080 -v ${pwd}/out:/app/out korche-placer
```

Each request will create a file:

```bash
out/placement-<pod-id>.json
```

---

### Changing the Listening Port

By default, the service listens on port `8080`.

To use a different port, you do not need to modify the code:  
simply map another external port to the container's internal port `8080`.

Example: expose the service on port `9090` instead of `8080`

```powerhell
docker run -p 9090:8080 korche-placer
```

Now the service will be reachable at:

```powershell
http://localhost:9090/place
http://<IP>:9090/place
```

## Calling the Service

```powershell
curl -X POST -H "Content-Type: application/json" --data-binary @request.json "http://<IP>:8080/place"
```

or, to specify the ouput as file rather than on STDOUT

```powershell
curl -X POST -H "Content-Type: application/json" --data-binary @request.json "http://localhost:8080/place?output=file"
```

Where `<IP>` can be:

- `localhost` (if you run the service locally)
- The IP address of the machine running the container
- A reachable address within your industrial network

---

## Health Check

The service exposes a dedicated health-check endpoint:

```bash
curl http://localhost:8080/health
```

If the service is running, it responds with:

```json
{"status":"ok"}
```

