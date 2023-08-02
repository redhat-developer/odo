const prismReactRenderer = require('prism-react-renderer');
const path = require('path');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'odo',
  tagline: 'odo - Fast iterative container-based application development',
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
            {
              label: 'YouTube',
              href: 'https://www.youtube.com/channel/UCXAt2CtoBBtN9EWe4xv4Row'
            }
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} odo Authors -- All Rights Reserved <br> Apache License 2.0 open source project`,
    },
    prism: {
      theme: prismReactRenderer.themes.github,
      darkTheme: prismReactRenderer.themes.oceanicNext,
      additionalLanguages: ['docker'],
    },
    algolia: {
      appId: '7RBQSTPIA4',
      apiKey: '97ac94cb47dcaeef1c2c9694bd39b458',
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
          exclude: [
              '**/docs-mdx/**',
              '**/_*.{js,jsx,ts,tsx,md,mdx}',
              '**/_*/**',
              '**/*.test.{js,jsx,ts,tsx}',
              '**/__tests__/**'
          ],
          versions: {
            current: {
              label: 'v3',
              badge: true,
              banner: 'none',
            },
            '2.5.0': {
              label: 'v2',
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
