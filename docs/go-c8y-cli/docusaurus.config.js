/** @type {import('@docusaurus/types').DocusaurusConfig} */

const path = require('path');
const isDev = process.env.NODE_ENV === 'development';

const isDeployPreview =
  process.env.NETLIFY && process.env.CONTEXT === 'deploy-preview';

const isBootstrapPreset = process.env.DOCUSAURUS_PRESET === 'bootstrap';

const baseUrl = process.env.BASE_URL || '/';

module.exports = {
  title: 'Cumulocity IoT CLI',
  tagline: 'Unofficial Cumulocity IoT Command Line Interface',
  url: 'https://reubenmiller.github.io/go-c8y-cli',
  baseUrl,
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'reubenmiller',
  projectName: 'go-c8y-cli',
  plugins: [
    '@docusaurus/plugin-google-analytics',
    [
      '@docusaurus/plugin-pwa',
      {
        debug: isDeployPreview,
        offlineModeActivationStrategies: [
          'appInstalled',
          'standalone',
          'queryString',
        ],
        swRegister: false,
        swCustom: path.resolve(__dirname, 'src/sw.js'),
        pwaHead: [
          {
            tagName: 'link',
            rel: 'manifest',
            href: `${baseUrl}manifest.json`,
          },
          {
            tagName: 'meta',
            name: 'theme-color',
            content: 'rgb(37, 194, 160)',
          },
          {
            tagName: 'meta',
            name: 'apple-mobile-web-app-capable',
            content: 'yes',
          },
          {
            tagName: 'meta',
            name: 'apple-mobile-web-app-status-bar-style',
            content: '#000',
          },
          {
            tagName: 'meta',
            name: 'msapplication-TileColor',
            content: '#000',
          },
        ],
      },
    ],
  ],
  themeConfig: {
    sidebarCollapsible: true,
    hideableSidebar: true,
    prism: {
      theme: require('prism-react-renderer/themes/github'),
      darkTheme: require('prism-react-renderer/themes/dracula'),
      additionalLanguages: ['powershell'],
    },
    googleAnalytics: {
      trackingID: 'UA-155263011-1',
      anonymizeIP: true,
    },
    announcementBar: {
      id: 'v2-major-release',
      content:
        '🎉 go-c8y-cli v2 is now supported natively in linux (pipelines and everything)! Check out the installation instructions',
        // '➡️ go-c8y-cli v2 is no longer only for PowerShell! <a target="_blank" rel="noopener noreferrer" href="https://v1.docusaurus.io/">v1.docusaurus.io</a>! 🔄',
    },
    navbar: {
      title: 'go-c8y-cli',
      logo: {
        alt: 'go-c8y-cli Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'doc',
          docId: 'demo',
          label: 'Docs',
          position: 'left',
        },
        {
          type: 'doc',
          docId: 'cli',
          label: 'API',
          position: 'left',
        },
        {to: 'blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/reubenmiller/go-c8y-cli',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        // {
        //   title: 'Community',
        //   items: [
        //     {
        //       label: 'Stack Overflow',
        //       href: 'https://stackoverflow.com/questions/tagged/docusaurus',
        //     },
        //     {
        //       label: 'Discord',
        //       href: 'https://discordapp.com/invite/docusaurus',
        //     },
        //     {
        //       label: 'Twitter',
        //       href: 'https://twitter.com/docusaurus',
        //     },
        //   ],
        // },
        {
          title: 'Community',
          items: [
            {
              label: 'Blog',
              to: 'blog/',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/reubenmiller/go-c8y-cli',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} go-c8y-cli. Built with Docusaurus.`,
    },
  },
  scripts: [
    {
      src: 'https://asciinema.org/a/326455.js',
      async: true,
    },
    // 'https://buttons.github.io/buttons.js',
    // 'https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.0/clipboard.min.js',
    // 'http://localhost:3000/js/code-block-buttons.js',
  ],
  stylesheets: [
    // 'http://localhost:3000/css/code-block-buttons.css'
  ],
  presets: [
    [
      isBootstrapPreset
        ? '@docusaurus/preset-bootstrap'
        : '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
            'https://github.com/reubenmiller/go-c8y-cli/edit/master/docs-next/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://github.com/facebook/docusaurus/edit/master/website/blog/',
        },
      },
    ],
  ],
};
