import { defineConfig } from 'vitepress'
import hairaGrammar from './haira.tmLanguage.json'

const sharedNav = (prefix: string) => [
  { text: prefix ? 'Docs' : 'Docs', link: `${prefix}/docs/getting-started/installation` },
  { text: 'Examples', link: `${prefix}/docs/examples` },
  { text: 'ARP', link: `${prefix}/agentic-rendering-protocol` },
  { text: 'Generative UI', link: `${prefix}/generative-ui` },
  {
    text: 'v0.3.0',
    items: [
      { text: 'Changelog', link: 'https://github.com/mrzdevcore/haira/releases' },
    ]
  }
]

const sharedSidebar = (prefix: string) => ({
  [`${prefix}/docs/`]: [
    {
      text: 'Getting Started',
      collapsed: false,
      items: [
        { text: 'Installation', link: `${prefix}/docs/getting-started/installation` },
        { text: 'Hello World', link: `${prefix}/docs/getting-started/hello-world` },
        { text: 'Key Concepts', link: `${prefix}/docs/getting-started/key-concepts` },
      ]
    },
    {
      text: 'Language Guide',
      collapsed: false,
      items: [
        { text: 'Variables & Types', link: `${prefix}/docs/language/variables-and-types` },
        { text: 'Functions', link: `${prefix}/docs/language/functions` },
        { text: 'Control Flow', link: `${prefix}/docs/language/control-flow` },
        { text: 'Structs & Enums', link: `${prefix}/docs/language/structs-and-enums` },
        { text: 'Error Handling', link: `${prefix}/docs/language/error-handling` },
        { text: 'Modules & Imports', link: `${prefix}/docs/language/modules` },
        { text: 'Pattern Matching', link: `${prefix}/docs/language/pattern-matching` },
        { text: 'Pipe Operator', link: `${prefix}/docs/language/pipe-operator` },
        { text: 'Methods', link: `${prefix}/docs/language/methods` },
        { text: 'Concurrency', link: `${prefix}/docs/language/concurrency` },
      ]
    },
    {
      text: 'Agentic',
      collapsed: false,
      items: [
        { text: 'Providers', link: `${prefix}/docs/agentic/providers` },
        { text: 'Tools', link: `${prefix}/docs/agentic/tools` },
        { text: 'Agents', link: `${prefix}/docs/agentic/agents` },
        { text: 'Workflows', link: `${prefix}/docs/agentic/workflows` },
        { text: 'Streaming', link: `${prefix}/docs/agentic/streaming` },
        { text: 'Agent Handoffs', link: `${prefix}/docs/agentic/handoffs` },
        { text: 'Memory & Sessions', link: `${prefix}/docs/agentic/memory` },
        { text: 'Generative UI', link: `${prefix}/docs/agentic/generative-ui` },
        { text: 'ARP Protocol', link: `${prefix}/docs/agentic/arp` },
      ]
    },
    {
      text: 'Standard Library',
      collapsed: true,
      items: [
        { text: 'Overview', link: `${prefix}/docs/stdlib/overview` },
        { text: 'HTTP & Server', link: `${prefix}/docs/stdlib/http` },
        { text: 'IO & File System', link: `${prefix}/docs/stdlib/io` },
        { text: 'JSON', link: `${prefix}/docs/stdlib/json` },
        { text: 'Postgres', link: `${prefix}/docs/stdlib/postgres` },
        { text: 'Strings', link: `${prefix}/docs/stdlib/strings` },
      ]
    },
    {
      text: 'Reference',
      collapsed: true,
      items: [
        { text: 'Examples', link: `${prefix}/docs/examples` },
        { text: 'Grammar', link: `${prefix}/docs/reference/grammar` },
        { text: 'Compiler Architecture', link: `${prefix}/docs/reference/compiler` },
      ]
    },
  ]
})

export default defineConfig({
  title: 'Haira',
  description: 'The programming language for AI agents and workflows',

  markdown: {
    languages: [hairaGrammar as any],
  },

  sitemap: {
    hostname: 'https://haira.dev',
  },

  head: [
    ['meta', { name: 'theme-color', content: '#E8A317' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'Haira' }],
    ['meta', { property: 'og:image', content: 'https://raw.githubusercontent.com/mrzdevcore/haira/main/assets/banner.svg' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: 'https://raw.githubusercontent.com/mrzdevcore/haira/main/assets/banner.svg' }],
    ['meta', { name: 'twitter:title', content: 'Haira — The programming language for AI agents' }],
    ['meta', { name: 'twitter:description', content: 'Build agents and workflows, not boilerplate. Four keywords. One binary.' }],
    ['meta', { name: 'author', content: 'mrzdevcore' }],
    ['meta', { name: 'keywords', content: 'haira, programming language, AI agents, workflows, LLM, generative UI, agentic, compiled language, Go' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', href: 'https://raw.githubusercontent.com/mrzdevcore/haira/main/assets/icon.svg' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', { href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap', rel: 'stylesheet' }],
    // Umami Analytics (website ID injected from UMAMI_WEBSITE_ID env var at build time)
    ...(process.env.UMAMI_WEBSITE_ID
      ? [['script', { defer: '', src: 'https://analytics.vzerolab.com/script.js', 'data-website-id': process.env.UMAMI_WEBSITE_ID }] as ['script', Record<string, string>]]
      : []),
    ['script', { type: 'application/ld+json' }, JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'SoftwareSourceCode',
      'name': 'Haira',
      'description': 'The programming language for AI agents and workflows. Build agents and workflows, not boilerplate.',
      'url': 'https://haira.dev',
      'codeRepository': 'https://github.com/mrzdevcore/haira',
      'programmingLanguage': 'Haira',
      'license': 'https://opensource.org/licenses/Apache-2.0',
      'author': {
        '@type': 'Person',
        'name': 'mrzdevcore',
        'url': 'https://github.com/mrzdevcore'
      }
    })],
  ],

  transformPageData(pageData) {
    const canonicalUrl = `https://haira.dev/${pageData.relativePath}`
      .replace(/index\.md$/, '')
      .replace(/\.md$/, '')

    pageData.frontmatter.head ??= []
    pageData.frontmatter.head.push(
      ['link', { rel: 'canonical', href: canonicalUrl }],
      ['meta', { property: 'og:url', content: canonicalUrl }],
    )

    if (pageData.frontmatter.title) {
      pageData.frontmatter.head.push(
        ['meta', { property: 'og:title', content: `${pageData.frontmatter.title} | Haira` }],
      )
    }
    if (pageData.frontmatter.description) {
      pageData.frontmatter.head.push(
        ['meta', { property: 'og:description', content: pageData.frontmatter.description }],
      )
    }
  },

  outDir: '../docs',
  cleanUrls: true,

  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: sharedNav('/zh'),
        sidebar: sharedSidebar('/zh'),
      }
    },
    ja: {
      label: '日本語',
      lang: 'ja',
      link: '/ja/',
      themeConfig: {
        nav: sharedNav('/ja'),
        sidebar: sharedSidebar('/ja'),
      }
    },
    ko: {
      label: '한국어',
      lang: 'ko',
      link: '/ko/',
      themeConfig: {
        nav: sharedNav('/ko'),
        sidebar: sharedSidebar('/ko'),
      }
    },
    es: {
      label: 'Español',
      lang: 'es',
      link: '/es/',
      themeConfig: {
        nav: sharedNav('/es'),
        sidebar: sharedSidebar('/es'),
      }
    },
    fr: {
      label: 'Français',
      lang: 'fr',
      link: '/fr/',
      themeConfig: {
        nav: sharedNav('/fr'),
        sidebar: sharedSidebar('/fr'),
      }
    },
  },

  themeConfig: {
    logo: 'https://raw.githubusercontent.com/mrzdevcore/haira/main/assets/icon.svg',
    siteTitle: false,

    nav: sharedNav(''),
    sidebar: sharedSidebar(''),

    socialLinks: [
      { icon: 'github', link: 'https://github.com/mrzdevcore/haira' }
    ],

    footer: {
      message: 'Released under the Apache-2.0 License.',
      copyright: 'Built with love by <a href="https://github.com/mrzdevcore" target="_blank">mrzdevcore</a>'
    },

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/mrzdevcore/haira/edit/main/website/:path',
      text: 'Edit this page on GitHub'
    }
  }
})
