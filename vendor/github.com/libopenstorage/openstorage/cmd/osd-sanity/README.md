# OpenStorage Test Command Line Program
This is the command line program that tests a OSD using the [`sanity`](https://github.com/libopenstorage/openstorage/tree/master/pkg/sanity) package test suite.

Example:

```
$ osd-sanity --osd.endpoint=<your osd server endpoint>
```

For verbose type:

```
$ osd-sanity --ginkgo.v --osd.endpoint=<your osd server endpoint>
```

### Help
The full Ginkgo and golang unit test parameters are available. Type

```
$ osd-sanity -h
```

to get more information

## Tutorial

Here is a model you can use to test `osd-sanity`. First make sure you are
running an NFS server on your system and have exported `/nfs`.

* Build osd: `make install`
* Run osd server on one terminal: `sudo $GOPATH/bin/osd -d -f etc/config/config.yaml`
* Download `osd-sanity` from releases or build: `cd cmd/osd-sanity && make`
* Run the test: `sudo ./osd-sanity --ginkgo.v --osd.endpoint=unix:///var/lib/osd/cluster/osd.sock`

## Download

Please see the [Releases](https://github.com/libopenstorage/openstorage/releases) page
to download the latest version of `osd-sanity`
