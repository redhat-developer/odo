const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'odo',
  tagline: 'Fast iterative Kubernetes and OpenShift development',
  url: 'https://odo.dev',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'openshift', // Usually your GitHub org/user name.
  projectName: 'odo', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'odo',
      // logo: {
      //   alt: 'My Site Logo',
      //   src: 'img/logo.svg',
      // },
      items: [
        {
          type: 'doc',
          docId: 'intro',
          position: 'left',
          label: 'Docs',
        },
        {to: '/blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/openshift/odo',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Learn',
          items: [
            {
              label: 'Installation',
              to: 'docs/getting-started/installation'
            },
          ]
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Slack',
              href: 'https://kubernetes.slack.com/archives/C01D6L2NUAG',
            },
            {
              label: 'Meetings',
              href: 'https://calendar.google.com/calendar/u/0/embed?src=gi0s0v5ukfqkjpnn26p6va3jfc@group.calendar.google.com',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'Blog',
              to: 'blog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/openshift/odo',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/rhdevelopers',
            },
          ],
        },
      ],
      // copyright: `Copyright © ${new Date().getFullYear()} My Project, Inc. Built with Docusaurus.`,
    },
    prism: {
      theme: lightCodeTheme,
      darkTheme: darkCodeTheme,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
            'https://github.com/openshift/odo/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://github.com/openshift/odo/edit/master/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
