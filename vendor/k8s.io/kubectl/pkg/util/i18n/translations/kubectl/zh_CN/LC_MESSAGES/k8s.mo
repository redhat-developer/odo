��    f      L  �   |      �  z   �  �   	  <   �	  S   
  <   b
  c  �
  �    .   �  "   �  4   
     ?     \    {  X   �  o   �    J  v   L  t   �  �  8  ;   �  [   9  J   �  a   �  �   B  �      �   �  %   u  W   �     �  u     4   �  -   �  3   �  2        Q  *   e  .   �  *   �  0   �  0     0   L  "   }     �  *   �  A   �     +  )   I     s     �      �  (   �     �  `     �   m  �   	     �     �  $   �     �       a   0  s   �  B     +   I  +   u  6   �  q   �  /   J   1   z   '   �      �   &   �   %   !  (   :!  #   c!      �!     �!  9   �!     "      "  #   :"  �   ^"  H   �"  &   *#  e   Q#  �   �#  E   �$  a   �$  �   E%  �   &     �&     �&  =   '  $   T'     y'  &   �'  +   �'     �'  r   (     t(  /   �(  �  �(  �   ;*  �   �*  8   H+  <   �+  6   �+  e  �+  �  [-  *   �.     
/  /   )/     Y/  !   y/    �/  _   �0  o   1    q1  v   s2  {   �2  �  f3  B   )5  [   l5  J   �5  a   6  �   u6  �   37  �   �7     �8  W   �8  &   9  u   F9  +   �9  $   �9  ,   :  X   ::     �:  %   �:  0   �:  .   ;  +   1;  *   ];  ,   �;     �;     �;  *   �;  E   <     d<  )   �<  !   �<     �<  )   �<  (   =     ==  `   T=  �   �=  �   Q>     �>     ?  (   #?     L?     g?  a   �?  s   �?  B   V@  +   �@  +   �@  6   �@  q   (A  (   �A  !   �A      �A     B  $   %B  '   JB  ,   rB  #   �B  '   �B     �B  =   C     FC     _C     yC  o   �C  9   	D     CD  Z   cD  �   �D  N   �E  R   �E  �   BF  �   G     �G     �G  =   �G     -H  "   IH  +   lH  +   �H     �H  f   �H     AI  (   UI     	   H       -   #                  3      `       d                  C          I       A          1           >   0          !           "   (       L   %       5   J   ?   4   )   b   Z   @      f   F       =         ;              c   ^         9   [   e   M      a   ,      S   '      \          Q             .   V   T   W       B      Y          E      6      :      X   &   /       P       D   K      U   2                      _   7   ]   <   8           R          $      G              O   N   *   
   +    
		  # Show metrics for all nodes
		  kubectl top node

		  # Show metrics for a given node
		  kubectl top node NODE_NAME 
		# Get the documentation of the resource and its fields
		kubectl explain pods

		# Get the documentation of a specific field of a resource
		kubectl explain pods.spec.containers 
		# Print flags inherited by all commands
		kubectl options 
		# Print the client and server versions for the current context
		kubectl version 
		# Print the supported API versions
		kubectl api-versions 
		# Show metrics for all pods in the default namespace
		kubectl top pod

		# Show metrics for all pods in the given namespace
		kubectl top pod --namespace=NAMESPACE

		# Show metrics for a given pod and its containers
		kubectl top pod POD_NAME --containers

		# Show metrics for the pods defined by label name=myLabel
		kubectl top pod -l name=myLabel 
		Convert config files between different API versions. Both YAML
		and JSON formats are accepted.

		The command takes filename, directory, or URL as input, and convert it into format
		of version specified by --output-version flag. If target version is not specified or
		not supported, convert to latest version.

		The default output will be printed to stdout in YAML format. One can use -o option
		to change to output destination. 
		Create a namespace with the specified name. 
		Create a role with single rule. 
		Create a service account with the specified name. 
		Mark node as schedulable. 
		Mark node as unschedulable. 
		Set the latest last-applied-configuration annotations by setting it to match the contents of a file.
		This results in the last-applied-configuration being updated as though 'kubectl apply -f <file>' was run,
		without updating any other parts of the object. 
	  # Create a new namespace named my-namespace
	  kubectl create namespace my-namespace 
	  # Create a new service account named my-service-account
	  kubectl create serviceaccount my-service-account 
	Create an ExternalName service with the specified name.

	ExternalName service references to an external DNS address instead of
	only pods, which will allow application authors to reference services
	that exist off platform, on other clusters, or locally. 
	Help provides help for any command in the application.
	Simply type kubectl help [path to command] for full details. 
    # Create a new LoadBalancer service named my-lbs
    kubectl create service loadbalancer my-lbs --tcp=5678:8080 
    # Dump current cluster state to stdout
    kubectl cluster-info dump

    # Dump current cluster state to /path/to/cluster-state
    kubectl cluster-info dump --output-directory=/path/to/cluster-state

    # Dump all namespaces to stdout
    kubectl cluster-info dump --all-namespaces

    # Dump a set of namespaces to /path/to/cluster-state
    kubectl cluster-info dump --namespaces default,kube-system --output-directory=/path/to/cluster-state 
    Create a LoadBalancer service with the specified name. A comma-delimited set of quota scopes that must all match each object tracked by the quota. A comma-delimited set of resource=quantity pairs that define a hard limit. A label selector to use for this budget. Only equality-based selector requirements are supported. A label selector to use for this service. Only equality-based selector requirements are supported. If empty (the default) infer the selector from the replication controller or replica set.) Additional external IP address (not managed by Kubernetes) to accept for the service. If this IP is routed to a node, the service can be accessed by this IP in addition to its generated service IP. An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field. Approve a certificate signing request Assign your own ClusterIP or set to 'None' for a 'headless' service (no loadbalancing). Attach to a running container ClusterIP to be assigned to the service. Leave empty to auto-allocate, or set to 'None' to create a headless service. ClusterRole this ClusterRoleBinding should reference ClusterRole this RoleBinding should reference Convert config files between different API versions Copy files and directories to and from containers. Create a TLS secret Create a namespace with the specified name Create a secret for use with a Docker registry Create a secret using specified subcommand Create a service account with the specified name Delete the specified cluster from the kubeconfig Delete the specified context from the kubeconfig Deny a certificate signing request Describe one or many contexts Display clusters defined in the kubeconfig Display merged kubeconfig settings or a specified kubeconfig file Display one or many resources Drain node in preparation for maintenance Edit a resource on the server Email for Docker registry Execute a command in a container Forward one or more local ports to a pod Help about any command If non-empty, set the session affinity for the service to this; legal values: 'None', 'ClientIP' If non-empty, the annotation update will only succeed if this is the current resource-version for the object. Only valid when specifying a single resource. If non-empty, the labels update will only succeed if this is the current resource-version for the object. Only valid when specifying a single resource. Mark node as schedulable Mark node as unschedulable Mark the provided resource as paused Modify certificate resources. Modify kubeconfig files Name or number for the port on the container that the service should direct traffic to. Optional. Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used. Output shell completion code for the specified shell (bash or zsh) Password for Docker registry authentication Path to PEM encoded public key certificate. Path to private key associated with given certificate. Precondition for resource version. Requires that the current resource version match this value in order to scale. Print the client and server version information Print the list of flags inherited by all commands Print the logs for a container in a pod Resume a paused resource Role this RoleBinding should reference Run a particular image on the cluster Run a proxy to the Kubernetes API server Server location for Docker registry Set specific features on objects Set the selector on a resource Show details of a specific resource or group of resources Show the status of the rollout Synonym for --target-port The image for the container to run. The image pull policy for the container. If left empty, this value will not be specified by the client and defaulted by the server The minimum number or percentage of available pods this budget requires. The name for the newly created object. The name for the newly created object. If not specified, the name of the input resource will be used. The name of the API generator to use. There are 2 generators: 'service/v1' and 'service/v2'. The only difference between them is that service port in v1 is named 'default', while it is left unnamed in v2. Default is 'service/v2'. The network protocol for the service to be created. Default is 'TCP'. The port that the service should serve on. Copied from the resource being exposed, if unspecified The resource requirement limits for this container.  For example, 'cpu=200m,memory=512Mi'.  Note that server side components may assign limits depending on the server configuration, such as limit ranges. The resource requirement requests for this container.  For example, 'cpu=100m,memory=256Mi'.  Note that server side components may assign requests depending on the server configuration, such as limit ranges. The type of secret to create Undo a previous rollout Update resource requests/limits on objects with pod templates Update the annotations on a resource Update the labels on a resource Update the taints on one or more nodes Username for Docker registry authentication View rollout history Where to output the files.  If empty or '-' uses stdout, otherwise creates a directory hierarchy in that directory dummy restart flag) kubectl controls the Kubernetes cluster manager Project-Id-Version: gettext-go-examples-hello
Report-Msgid-Bugs-To: EMAIL
PO-Revision-Date: 2017-11-11 19:01+0800
Last-Translator: zhengjiajin <zhengjiajin@caicloud.io>
Language-Team: 
Language: zh
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit
X-Generator: Poedit 2.0.4
X-Poedit-SourceCharset: UTF-8
Plural-Forms: nplurals=2; plural=(n > 1);
 
		  # 显示所有 nodes 上的指标
		  kubectl top node

		  # 显示指定 node 上的指标
		  kubectl top node NODE_NAME 
		# 获取资源及其字段的文档
		kubectl explain pods

		# 获取资源指定字段的文档
		kubectl explain pods.spec.containers 
		# 输出所有命令继承的 flags
		kubectl options 
		# 输出当前 client 和 server 版本
		kubectl version 
		# 输出支持的 API 版本
		kubectl api-versions 
		# 显示 default namespace 下所有 pods 下的 metrics
		kubectl top pod

		# 显示指定 namespace 下所有 pods 的 metrics
		kubectl top pod --namespace=NAMESPACE

		# 显示指定 pod 和它的容器的 metrics
		kubectl top pod POD_NAME --containers

		# 显示指定 label 为 name=myLabel 的 pods 的 metrics
		kubectl top pod -l name=myLabel 
		在不同的 API versions 转换配置文件. 接受 YAML
		和 JSON 格式.

		这个命令以 filename, directory, 或者 URL 作为输入, 并通过 —output-version flag
		 转换到指定版本的格式. 如果目标版本没有被指定或者
		不支持, 转换到最后的版本.

		默认以 YAML 格式输出到 stdout. 可以使用 -o option
		修改目标输出的格式. 
		创建一个 namespace 并指定名称. 
		创建单一 rule 的 role. 
		创建一个指定名称的 service account. 
		标记 node 为 schedulable. 
		标记 node 为 unschedulable. 
		Set the latest last-applied-configuration annotations by setting it to match the contents of a file.
		This results in the last-applied-configuration being updated as though 'kubectl apply -f <file>' was run,
		without updating any other parts of the object. 
	  # 创建一个名称为 my-namespace 的 namespace
	  kubectl create namespace my-namespace 
	  # Create a new service account named my-service-account
	  kubectl create serviceaccount my-service-account 
	Create an ExternalName service with the specified name.

	ExternalName service references to an external DNS address instead of
	only pods, which will allow application authors to reference services
	that exist off platform, on other clusters, or locally. 
	Help provides help for any command in the application.
	Simply type kubectl help [path to command] for full details. 
    # 创建一个名称为 my-lbs 的 LoadBalancer service
    kubectl create service loadbalancer my-lbs --tcp=5678:8080 
    # 导出当前的集群状态信息到 stdout
    kubectl cluster-info dump

    # 导出当前的集群状态 /path/to/cluster-state
    kubectl cluster-info dump --output-directory=/path/to/cluster-state

    # 导出所有分区到 stdout
    kubectl cluster-info dump --all-namespaces

    # 导出一组分区到 /path/to/cluster-state
    kubectl cluster-info dump --namespaces default,kube-system --output-directory=/path/to/cluster-state 
    使用一个指定的名称创建一个 LoadBalancer service. A comma-delimited set of quota scopes that must all match each object tracked by the quota. A comma-delimited set of resource=quantity pairs that define a hard limit. A label selector to use for this budget. Only equality-based selector requirements are supported. A label selector to use for this service. Only equality-based selector requirements are supported. If empty (the default) infer the selector from the replication controller or replica set.) Additional external IP address (not managed by Kubernetes) to accept for the service. If this IP is routed to a node, the service can be accessed by this IP in addition to its generated service IP. An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field. 同意一个自签证书请求 Assign your own ClusterIP or set to 'None' for a 'headless' service (no loadbalancing). Attach 到一个运行中的 container ClusterIP to be assigned to the service. Leave empty to auto-allocate, or set to 'None' to create a headless service. ClusterRoleBinding 应该指定 ClusterRole RoleBinding 应该指定 ClusterRole 在不同的 API versions 转换配置文件 复制 files 和 directories 到 containers 和从容器中复制 files 和 directories. 创建一个 TLS secret 创建一个指定名称的 namespace 创建一个给 Docker registry 使用的 secret 使用指定的 subcommand 创建一个 secret 创建一个指定名称的 service account 删除 kubeconfig 文件中指定的集群 删除 kubeconfig 文件中指定的 context 拒绝一个自签证书请求 描述一个或多个 contexts 显示 kubeconfig 文件中定义的集群 显示合并的 kubeconfig 配置或一个指定的 kubeconfig 文件 显示一个或更多 resources Drain node in preparation for maintenance 在服务器上编辑一个资源 Email for Docker registry 在一个 container 中执行一个命令 Forward one or more local ports to a pod Help about any command If non-empty, set the session affinity for the service to this; legal values: 'None', 'ClientIP' If non-empty, the annotation update will only succeed if this is the current resource-version for the object. Only valid when specifying a single resource. If non-empty, the labels update will only succeed if this is the current resource-version for the object. Only valid when specifying a single resource. 标记 node 为 schedulable 标记 node 为 unschedulable 标记提供的 resource 为中止状态 修改 certificate 资源. 修改 kubeconfig 文件 Name or number for the port on the container that the service should direct traffic to. Optional. Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used. Output shell completion code for the specified shell (bash or zsh) Password for Docker registry authentication Path to PEM encoded public key certificate. Path to private key associated with given certificate. Precondition for resource version. Requires that the current resource version match this value in order to scale. 输出 client 和 server 的版本信息 输出所有命令的层级关系 输出容器在 pod 中的日志 继续一个停止的 resource RoleBinding 的 Role 应该被引用 在集群中运行一个指定的镜像 运行一个 proxy 到 Kubernetes API server Server location for Docker registry 为 objects 设置一个指定的特征 设置 resource 的 selector 显示一个指定 resource 或者 group 的 resources 详情 显示 rollout 的状态 Synonym for --target-port 指定容器要运行的镜像. 容器的镜像拉取策略. 如果为空, 这个值将不会 被 client 指定且使用 server 端的默认值 最小数量百分比可用的 pods 作为 budget 要求. 名称为最新创建的对象. 名称为最新创建的对象. 如果没有指定, 输入资源的 名称即将被使用. 使用 generator 的名称. 这里有 2 个 generators: 'service/v1' 和 'service/v2'. 为一个不同地方是服务端口在 v1 的情况下叫 'default', 如果在 v2 中没有指定名称. 默认的名称是 'service/v2'. 创建 service 的时候伴随着一个网络协议被创建. 默认是 'TCP'. 服务的端口应该被指定. 如果没有指定, 从被创建的资源中复制 The resource requirement limits for this container.  For example, 'cpu=200m,memory=512Mi'.  Note that server side components may assign limits depending on the server configuration, such as limit ranges. 资源为 container 请求 requests . 例如, 'cpu=100m,memory=256Mi'. 注意服务端组件也许会赋予 requests, 这决定于服务器端配置, 比如 limit ranges. 创建 secret 类型资源 撤销上一次的 rollout 在对象的 pod templates 上更新资源的 requests/limits 更新一个资源的注解 更新在这个资源上的 labels 更新一个或者多个 node 上的 taints Username 为 Docker registry authentication 显示 rollout 历史 输出到 files.  如果是 empty or '-' 使用 stdout, 否则创建一个 目录层级在那个目录 dummy restart flag) kubectl 控制 Kubernetes cluster 管理 