const path = require('path');

module.exports = function (context, options) {
    const isProd = process.env.NODE_ENV === 'production';

    // The following content can be obtained from https://segment.com/docs/connections/sources/catalog/libraries/website/javascript/quickstart/
    // It has been modified to anonymize IP address of the user;
    // for more details, refer https://segment.com/docs/connections/sources/catalog/libraries/website/javascript/identity/#anonymizing-ip.
    const contents = `!function(){var analytics=window.analytics=window.analytics||[];if(!analytics.initialize)if(analytics.invoked)window.console&&console.error&&console.error("Segment snippet included twice.");else{analytics.invoked=!0;analytics.methods=["trackSubmit","trackClick","trackLink","trackForm","pageview","identify","reset","group","track","ready","alias","debug","page","once","off","on","addSourceMiddleware","addIntegrationMiddleware","setAnonymousId","addDestinationMiddleware"];analytics.factory=function(e){return function(){var t=Array.prototype.slice.call(arguments);t.unshift(e);analytics.push(t);return analytics}};for(var e=0;e<analytics.methods.length;e++){var key=analytics.methods[e];analytics[key]=analytics.factory(key)}analytics.load=function(key,e){var t=document.createElement("script");t.type="text/javascript";t.async=!0;t.src="https://cdn.segment.com/analytics.js/v1/" + key + "/analytics.min.js";var n=document.getElementsByTagName("script")[0];n.parentNode.insertBefore(t,n);analytics._loadOptions=e};analytics._writeKey="seYXMF0tyHs5WcPsaNXtSEmQk3FqzTz0";;analytics.SNIPPET_VERSION="4.15.3";
  analytics.load("seYXMF0tyHs5WcPsaNXtSEmQk3FqzTz0");
  analytics.page({}, {context: {ip: '0.0.0.0'}});
  }}();`

    return {
        name: 'docusaurus-odo-plugin-segment',

        getClientModules() {
            return isProd ? [path.resolve(__dirname, './segment')] : [];
        },

        injectHtmlTags() {
            if (!isProd) {
                return {};
            }
            return {
                headTags: [
                    {
                        tagName: 'link',
                        attributes: {
                            rel: 'preconnect',
                            href: 'https://cdn.segment.io',
                        },
                    },
                    {
                        tagName: 'script',
                        innerHTML: contents + '\n',
                    },
                ],
            };
        },
    };
};
