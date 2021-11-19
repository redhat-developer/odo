import ExecutionEnvironment from "@docusaurus/ExecutionEnvironment";

export default (function () {
    if (!ExecutionEnvironment.canUseDOM) {
        return null;
    }

    return {
        onRouteUpdate({ location }) {
            if (!window.analytics) return;
            // Call to `page` has been modified to anonymize IP address of the user;
            // for more details, refer https://segment.com/docs/connections/sources/catalog/libraries/website/javascript/identity/#anonymizing-ip.
            window.analytics.page({}, {context: {ip: '0.0.0.0'}});
        },
    };
})();
