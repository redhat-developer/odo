# All constructs and no magic

The artifacts here are purely hand written and has use no magic provided by current implementation of kedge.

## Deploy

Deploying application in this requires secrets to be created already:

```bash
oc create secret generic wordpress --from-literal='MYSQL_ROOT_PASSWORD=rootpasswd,DB_PASSWD=wordpress'
kedge generate -f web.yaml -f db.yaml | oc create -f -
```

