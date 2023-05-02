# Semantic of commands

Components:
- container
- cluster resource (Kubernetes/OpenShift)
- volume
- image

| Command                     | PreStart | PostStart | PreStop | PostStop  |
|-----------------------------|----------|-----------|---------|-----------|
| exec on container           |          |   Y(1)    |         |           |
| exec on cluster resource    |  N/A     |   N/A     |  N/A    |   N/A     |
| exec on volume              |  N/A     |   N/A     |  N/A    |   N/A     |
| exec on image               |  N/A     |   N/A     |  N/A    |   N/A     |
| &nbsp;                      |          |           |         |           |
| apply on container          |    ?     |     ?     |   ?     |    ?      |
| apply on cluster resource   |          |     0     |         |           |
| apply on volume             |    ?     |     ?     |   ?     |    ?      |
| apply on image              |          |     0     |         |           |
| &nbsp;                      |          |           |         |           |
| composite serial            |          |           |         |           |
| composite parallel          |          |           |         |           |


| Command                     | Build | Run/Debug | Deploy |
|-----------------------------|-------|-----------|--------|
| exec on container           | Y(1)  |   Y(1)    |  Y(2)  |
| exec on cluster resource    | N/A   |   N/A     |  N/A   |
| exec on volume              | N/A   |   N/A     |  N/A   |
| exec on image               | N/A   |   N/A     |  N/A   |
| &nbsp;                      |       |           |        |
| apply on container          |   ?   |     ?     |    ?   |
| apply on cluster resource   |   0   |     Y     |    Y   |
| apply on volume             |   ?   |     ?     |    ?   |
| apply on image              |   0   |     Y     |    Y   |
| &nbsp;                      |       |           |        |
| composite serial            |       |           |        |
| composite parallel          |       |           |        |


Legend:

- 0: Supported by handler but not implemented
- Y: Implemented
- N/A: Not applicable (by spec)
- ?: Spec is not clear

- (1) Implemented in pkg/component
- (2) Implemented in pkg/deploy
