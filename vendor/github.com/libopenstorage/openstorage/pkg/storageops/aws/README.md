### To test AWS

You will first need to create an AWS instance and then provide details of this instance as below.

```bash
export AWS_REGION=<instance-region>
export AWS_ZONE=<aws-availibility-zone>
export AWS_INSTANCE_NAME=<aws-instance-name>
export AWS_ACCESS_KEY_ID=<access-id>
export AWS_SECRET_ACCESS_KEY=<access-secret>
go test
```
