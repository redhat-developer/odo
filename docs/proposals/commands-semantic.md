# Semantic of commands

Components:
- container
- cluster resource (Kubernetes/OpenShift)
- volume
- image

| Command                     | PreStart | PostStart | PreStop | PostStop  |
|-----------------------------|----------|-----------|---------|-----------|
| exec on container           |          |    Yt     |   Yt    |           |
| exec on cluster resource    |  N/A     |   N/A     |  N/A    |   N/A     |
| exec on volume              |  N/A     |   N/A     |  N/A    |   N/A     |
| exec on image               |  N/A     |   N/A     |  N/A    |   N/A     |
| &nbsp;                      |          |           |         |           |
| apply on container          |    ?     |     ?     |   ?     |    ?      |
| apply on cluster resource   |          |    Yt     |   Yt    |           |
| apply on volume             |    ?     |     ?     |   ?     |    ?      |
| apply on image              |          |    Yt     |   Yt    |           |
| &nbsp;                      |          |           |         |           |
| composite serial            |          |           |         |           |
| composite parallel          |          |           |         |           |


| Command                     | Build | Run/Debug | Deploy |
|-----------------------------|-------|-----------|--------|
| exec on container           |  Yt   |    Yt     |   Yt   |
| exec on cluster resource    | N/A   |   N/A     |  N/A   |
| exec on volume              | N/A   |   N/A     |  N/A   |
| exec on image               | N/A   |   N/A     |  N/A   |
| &nbsp;                      |       |           |        |
| apply on container          |   ?   |     ?     |    ?   |
| apply on cluster resource   |  Yt   |    Yt     |   Yt   |
| apply on volume             |   ?   |     ?     |    ?   |
| apply on image              |  Yt   |    Yt     |   Yt   |
| &nbsp;                      |       |           |        |
| composite serial            |       |           |        |
| composite parallel          |       |           |        |


Legend:

- 0: Supported by handler but not implemented
- Y: Implemented by pkg/component.NewRunHandler (Yt: tested in pkg/component/handler_test.go)
- N/A: Not applicable (by spec)
- ?: Spec is not clear
