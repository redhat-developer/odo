# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

#-----------------------------------------------------------------------------
# VERBOSE target
#-----------------------------------------------------------------------------

# When you run make VERBOSE=1 (the default), executed commands will be printed
# before executed. If you run make VERBOSE=2 verbose flags are turned on and
# quiet flags are turned off for various commands. Use V_FLAG in places where
# you can toggle on/off verbosity using -v. Use Q_FLAG in places where you can
# toggle on/off quiet mode using -q. Use S_FLAG where you want to toggle on/off
# silence mode using -s...
VERBOSE ?= 1
Q = @
Q_FLAG = -q
QUIET_FLAG = --quiet
V_FLAG =
VERBOSE_FLAG =
S_FLAG = -s
X_FLAG =
ifeq ($(VERBOSE),1)
	Q =
endif
ifeq ($(VERBOSE),2)
	Q =
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	VERBOSE_FLAG = --verbose
	X_FLAG = -x
endif

#-----------------------------------------------------------------------------
# Examples Commons
#-----------------------------------------------------------------------------
EC=$(SHELL) -c '. ../../hack/examples-commons.sh && $$1' EC

export HACK_YAMLS=../../hack/yamls

## -- Commmon Utility targets --

## Print help message for all Makefile targets
## Run `make` or `make help` to see the help
.PHONY: help
help: ## Credit: https://gist.github.com/prwhite/8168133#gistcomment-2749866

	@printf "Usage:\n  make <target>";

	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                     "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                     "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)

## -- Common Cluster Admin Targets --

## --- Service Binding Operator ---

## ---- Community version ----

.PHONY: install-service-binding-operator-subscription-community
## Install the Service Binding Operator Subscription
install-service-binding-operator-subscription-community:
	${Q}${EC} install_service_binding_operator_subscription_community

.PHONY: install-service-binding-operator-community
## Install the Service Binding Operator
install-service-binding-operator-community: install-service-binding-operator-subscription-community

.PHONY: uninstall-service-binding-operator-subscription-community
## Uninstall the Service Binding Operator Subscription
uninstall-service-binding-operator-subscription-community:
	${Q}${EC} uninstall_service_binding_operator_subscription_community

.PHONY: uninstall-service-binding-operator-community
## Uninstall the Service Binding Operator
uninstall-service-binding-operator-community: uninstall-service-binding-operator-subscription-community

## ---- Latest master ----

.PHONY: install-service-binding-operator-source-master
## Install the Service Binding Operator Source for latest master
install-service-binding-operator-source-master:
	${Q}${EC} install_service_binding_operator_source_master

.PHONY: uninstall-service-binding-operator-source-master
## Uninstall the Service Binding Operator Source for latest master
uninstall-service-binding-operator-source-master:
	${Q}${EC} uninstall_service_binding_operator_source_master

.PHONY: install-service-binding-operator-subscription-master
## Install the Service Binding Operator Subscription for latest master
install-service-binding-operator-subscription-master:
	${Q}${EC} install_service_binding_operator_subscription_master

.PHONY: install-service-binding-operator-master
## Install the Service Binding Operator for latest master
install-service-binding-operator-master: install-service-binding-operator-source-master install-service-binding-operator-subscription-master

.PHONY: uninstall-service-binding-operator-subscription-master
## Uninstall the Service Binding Operator Subscription for latest master
uninstall-service-binding-operator-subscription-master:
	${Q}${EC} uninstall_service_binding_operator_subscription_master

.PHONY: uninstall-service-binding-operator-master
## Uninstall the Service Binding Operator for latest master
uninstall-service-binding-operator-master: uninstall-service-binding-operator-subscription-master uninstall-service-binding-operator-source-master

## --- Backing Service DB (PostgreSQL) Operator ---

.PHONY: install-backing-db-operator-source
## Install the Backing Service DB Operator Source
install-backing-db-operator-source:
	${Q}${EC} install_postgresql_operator_source

.PHONY: install-backing-db-operator-subscription
## Install the Backing Service DB Operator Subscription
install-backing-db-operator-subscription:
	${Q}${EC} install_postgresql_operator_subscription

.PHONY: install-backing-db-operator
## Install the Backing Service DB Operator
install-backing-db-operator: install-backing-db-operator-source install-backing-db-operator-subscription

.PHONY: uninstall-backing-db-operator-source
## Uninstall the Backing Service DB Operator Source
uninstall-backing-db-operator-source:
	${Q}${EC} uninstall_postgresql_operator_source

.PHONY: uninstall-backing-db-operator-subscription
## Uninstall the Backing Service DB Operator Subscription
uninstall-backing-db-operator-subscription:
	${Q}${EC} uninstall_postgresql_operator_subscription

.PHONY: uninstall-backing-db-operator
## Uninstall the Backing Service DB Operator
uninstall-backing-db-operator: uninstall-backing-db-operator-subscription uninstall-backing-db-operator-source

## --- Serverless Operator ---

.PHONY: install-serverless-operator-subscription
## Install the Serverless Operator Subscription
install-serverless-operator-subscription:
	${Q}${EC} install_serverless_operator_subscription

.PHONY: install-serverless-operator
## Install the Serverless Operator
install-serverless-operator: install-serverless-operator-subscription

.PHONY: uninstall-serverless-operator-subscription
## Uninstall the Serverless Operator Subscription
uninstall-serverless-operator-subscription:
	${Q}${EC} uninstall_serverless_operator_subscription

.PHONY: uninstall-serverless-operator
## Uninstall the Serverless Operator
uninstall-serverless-operator: uninstall-serverless-operator-subscription

## --- Service Mesh Operator ---

.PHONY: install-service-mesh-operator-subscription
## Install the Service Mesh Operator Subscription
install-service-mesh-operator-subscription:
	${Q}${EC} install_service_mesh_operator_subscription

.PHONY: install-service-mesh-operator
## Install the Service Mesh Operator
install-service-mesh-operator: install-service-mesh-operator-subscription

.PHONY: uninstall-service-mesh-operator-subscription
## Uninstall the Service Mesh Operator Subscription
uninstall-service-mesh-operator-subscription:
	${Q}${EC} uninstall_service_mesh_operator_subscription

.PHONY: uninstall-service-mesh-operator
## Uninstall the Service Mesh Operator
uninstall-service-mesh-operator: uninstall-service-mesh-operator-subscription

## --- Knative Serving (Serverless UI)

.PHONY: install-knative-serving
## Install Knative Serving
install-knative-serving:
	${Q}${EC} install_knative_serving

.PHONY: uninstall-knative-serving
## Uninstall Knative Serving
uninstall-knative-serving:
	${Q}-${EC} uninstall_knative_serving

# === Quarkus Native S2i Buider Image ===

.PHONY: install-quarkus-native-s2i-builder
## Install ubi-quarkus-native-s2i builder
install-quarkus-native-s2i-builder:
	${Q}${EC} install_ubi_quarkus_native_s2i_builder_image

## -- Common Application Developer Targets --

.PHONY: create-project
## Create the OpenShift project/namespace
create-project:
	${Q}-${EC} create_project

.PHONY: delete-project
## Delete the OpenShift project/namespace
delete-project:
	${Q}${EC} delete_project

.PHONY: create-backing-db-instance
## Create the Backing Service DB Operator
create-backing-db-instance:
	${Q}${EC} install_postgresql_db_instance
