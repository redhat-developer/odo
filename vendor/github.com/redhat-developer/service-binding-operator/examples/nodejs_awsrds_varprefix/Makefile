include ../commons.mk

.DEFAULT_GOAL := help

## -- Cluster Admin targets --

.PHONY: install-aws-rds-operator-source
## Install the AWS RDS operator
install-aws-rds-operator-source:
	${Q}oc apply -f operator-source.aws-rds-operator.yaml

.PHONY: install-operator-secrets
## Install the AWS RDS operator secrets
install-operator-secrets: install-aws-rds-operator-secrets

.PHONY: install-aws-rds-operator-secrets
## Install the AWS RDS operator secrets
install-aws-rds-operator-secrets:
	$(eval AWS_ACCESS_KEY_ID := $(shell cat ~/.aws/credentials | grep aws_access_key_id | cut -d " " -f3))
	$(eval AWS_SECRET_ACCESS_KEY := $(shell cat ~/.aws/credentials | grep aws_secret_access_key | cut -d " " -f3))
	${Q}sed -e 's,REPLACE_AWS_ACCESS_KEY_ID,$(shell echo -n "$(AWS_ACCESS_KEY_ID)" | base64),g' secret.aws-rds-operator.yaml | \
	 sed -e 's,REPLACE_AWS_SECRET_ACCESS_KEY,$(shell echo -n "$(AWS_SECRET_ACCESS_KEY)" | base64),g' | oc apply -f -

## -- Application Developer targets --

.PHONY: deploy-shell-app
## Deploy the shell application
deploy-shell-app:
	${Q}cat image-stream.shell-app.yaml build-config.shell-app.yaml deployment-config.shell-app.yaml | oc apply -f -

.PHONY: deploy-mydb
## Deploy the database that supports the applications
deploy-mydb:
	${Q}cat secret.mydb.yaml rds.mydb.yaml | oc apply -f -

.PHONY: set-labels-on-nodejs-app
## Set the binding labels for the NodeJS app
set-labels-on-nodejs-app:
	${Q}oc patch dc nodejs-app -p '{"metadata": {"labels": {"connects-to": "postgres", "environment": "nodejs"}}}'

.PHONY: create-service-binding-request-shell-app
## Create the Service Binding Request for the shell app
create-service-binding-request-shell-app:
	${Q}oc apply -f service-binding-request.shell-app.yaml

.PHONY: create-service-binding-request-nodejs-app
## Create the Service Binding Request for the NodeJS app
create-service-binding-request-nodejs-app:
	${Q}oc apply -f service-binding-request.nodejs-app.yaml
