
# `odo list endpoint`
List endpoints defined in local devfile and in cluster

## flags
- `-o json` Output information in json format


## example

```
$ odo list endpoints
Found the following URLs for component devfile-nodejs-deploy
NAME            URL                       PORT     SECURE     KIND
http-3000       <provided by cluster>     3000     false      route
```

The output is similar to v2 `odo url list` output. The only difference is that odo does't show `STATE` anymore.
The reason is that `endpoints` are mainly used in inner loop style of deployment and with changes introduced in `odo dev` command (resources are deleted automaticaly when command is not running) the `STATE` doesn't make much sense anymore