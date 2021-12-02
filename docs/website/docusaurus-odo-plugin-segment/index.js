const path = require('path');
const snippet = require('@segment/snippet');

module.exports = function (context, fromOptions) {
    const isProd = process.env.NODE_ENV === 'production';
    const {siteConfig} = context;
    const {themeConfig} = siteConfig;
    const {segment: fromThemeConfig} = themeConfig || {};

    const segment = {
        ...fromThemeConfig,
        ...fromOptions
    };

    const {apiKey} = segment;
    const {options} = segment;

    if (!apiKey) {
        throw new Error('Unable to find a Segment `apiKey` in `plugin` options or `themeConfig`.');
    }

    let snippetContent = snippet.min(segment);
    // Modify the page() call to anonymize IP address of the user;
    let pageCall = "analytics.page({}, " + JSON.stringify(options) + ")"
    // Replace the page function with page({}, {context: {ip: '0.0.0.0'}})
    const contents = snippetContent.replace("analytics.page()", pageCall)
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
