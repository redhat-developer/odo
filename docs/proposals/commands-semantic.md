# Semantic of commands

Components:
- container
- cluster resource (Kubernetes/OpenShift)
- volume
- image

| Command                     | PreStart | PostStart | PreStop | PostStop  |
|-----------------------------|----------|-----------|---------|-----------|
| exec on container           |          |   Y(1)    |         |           |
| exec on cluster resource    |          |           |         |           |
| exec on volume              |          |           |         |           |
| exec on image               |          |           |         |           |
| &nbsp;                      |          |           |         |           |
| apply on container          |          |           |         |           |
| apply on cluster resource   |          |     0     |         |           |
| apply on volume             |          |           |         |           |
| apply on image              |          |     0     |         |           |
| &nbsp;                      |          |           |         |           |
| composite                   |          |           |         |           |


| Command                     | Build | Run/Debug | Deploy |
|-----------------------------|-------|-----------|--------|
| exec on container           | Y(1)  |   Y(1)    |  Y(2)  |
| exec on cluster resource    |       |           |        |
| exec on volume              |       |           |        |
| exec on image               |       |           |        |
| &nbsp;                      |       |           |        |
| apply on container          |       |           |        |
| apply on cluster resource   |   0   |     Y     |    Y   |
| apply on volume             |       |           |        |
| apply on image              |   0   |     Y     |    Y   |
| &nbsp;                      |       |           |        |
| composite                   |       |           |        |


0: Supported by handler but not implemented
Y: Implemented

(1) Implemented in pkg/component
(2) Implemented in pkg/deploy
