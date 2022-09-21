const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/oceanicNext');
const path = require('path');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'odo',
  tagline: 'odo - Fast iterative Kubernetes and OpenShift development',
  url: 'https://odo.dev',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'redhat-developer', // Usually your GitHub org/user name.
  projectName: 'odo', // Usually your repo name.
  plugins: [
    [
      path.resolve(__dirname, 'docusaurus-odo-plugin-segment'),
      {
        apiKey: 'seYXMF0tyHs5WcPsaNXtSEmQk3FqzTz0',
        options: {
          context: { ip: '0.0.0.0' }
        }
      }
    ]
  ],
  themeConfig: {
    docs: {
      sidebar: {
        autoCollapseCategories: false
      },
    },
    announcementBar: {
      id: 'announcementBar-2', // Increment on change
      content: `⭐️ Love odo? Support us by giving it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/redhat-developer/odo">GitHub</a>! ⭐️`,
    },
    navbar: {
      title: 'odo',
      logo: {
        alt: 'odo Logo',
        src: 'img/logo.png',
        srcDark: 'img/logo_dark.png',
      },
      items: [
        {
          type: 'doc',
          docId: 'introduction',
          position: 'left',
          label: 'Docs',
        },
        { to: '/blog', label: 'Blog', position: 'left' },
        {
          href: 'https://github.com/redhat-developer/odo',
          label: 'GitHub',
          position: 'right',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
          dropdownActiveClassDisabled: true,
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
              to: 'docs/overview/installation'
            },
            {
              label: 'Quickstart',
              to: 'docs/user-guides/quickstart'
            },
          ]
        },
        {
          title: 'Community',
          items: [
            {
              label: '#odo on the Kubernetes Slack',
              href: 'https://slack.k8s.io/',
              external: true,
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
              href: 'https://github.com/redhat-developer/odo',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/rhdevelopers',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} odo Authors -- All Rights Reserved <br> Apache License 2.0 open source project`,
    },
    prism: {
      theme: lightCodeTheme,
      darkTheme: darkCodeTheme,
    },
    algolia: {
      appId: 'BH4D9OD16A',
      apiKey: 'e498f97159ee3094d356b8ed95dd405f',
      indexName: 'odo',
      debug: false
    }
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          breadcrumbs: true,
          sidebarCollapsible: true,
          lastVersion: 'current',
          versions: {
            current: {
              label: '3.0.0 (RC 1) 🚧',
              badge: true,
              banner: 'none',
            },
            '2.5.0': {
              label: '2.5.0 (Stable) 🚀',
              path: '2.5.0',
              badge: true,
              banner: 'none',
            },
          },
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
            'https://github.com/redhat-developer/odo/edit/main/docs/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://github.com/redhat-developer/odo/edit/main/docs/website/',
          blogSidebarTitle: 'All posts',
          blogSidebarCount: 'ALL',
          postsPerPage: 5,
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
