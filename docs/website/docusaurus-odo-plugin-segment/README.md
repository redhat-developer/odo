This plugin has been derived from https://github.com/xer0x/docusaurus-plugin-segment.

Refer https://docusaurus.io/docs/using-plugins for more information on using, and configuring a docusaurus plugin.

To install this plugin, run the following command:
```shell
cd $ODOPATH/docs/website
npm install --save ./docusaurus-odo-plugin-segment
```

_This plugin was written for the following reason_:

All the other plugins use https://github.com/segmentio/snippet to load analytics.js data. The library does not support `options` argument required by `page`, `track`, and `identify` and other calls to accept certain data, in our case, it is required to anonymize user IP.

Following is an example of `page` event which can anonymize user IP.
```go
analytics.page({}, {context: ip: {'0.0.0.0'}})
```

Refer [Basic Tracking Methods](https://segment.com/docs/connections/sources/catalog/libraries/website/javascript/#basic-tracking-methods) to see all the supported events and their arguments.

Since we cannot pass `options` argument to any of the plugin, it is not possible to anonymize user IP, because of which we decided to write our own plugin.

**NOTE:**
This plugin is specific to odo, and this directory within the odo repo is the source of truth; we do not maintain a separate git repo for this plugin. This plugin is not to be used for any other project, at least until we figure out a way to retrieve the Write Key from the docusaurus plugin config, and add the `context` part to `page()` calls like discussed in the previous section.
