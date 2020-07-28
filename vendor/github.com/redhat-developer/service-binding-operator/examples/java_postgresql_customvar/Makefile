.DEFAULT_GOAL := help

include ../commons.mk

## -- Application Developer targets --

.PHONY: create-service-binding-request
## Create the Service Binding Request
create-service-binding-request:
	${Q}oc apply -f service-binding-request.yaml
