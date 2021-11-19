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
