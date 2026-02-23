import { defineConfig } from 'vitepress'
import hairaGrammar from './haira.tmLanguage.json'

export default defineConfig({
  title: 'Haira',
  description: 'The programming language for AI agents and workflows',
  lang: 'en-US',

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

  themeConfig: {
    logo: 'https://raw.githubusercontent.com/mrzdevcore/haira/main/assets/icon.svg',
    siteTitle: false,

    nav: [
      { text: 'Docs', link: '/docs/getting-started/installation' },
      { text: 'Examples', link: '/docs/examples' },
      { text: 'ARP', link: '/agentic-rendering-protocol' },
      { text: 'Generative UI', link: '/generative-ui' },
      {
        text: 'v0.3.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/mrzdevcore/haira/releases' },
        ]
      }
    ],

    sidebar: {
      '/docs/': [
        {
          text: 'Getting Started',
          collapsed: false,
          items: [
            { text: 'Installation', link: '/docs/getting-started/installation' },
            { text: 'Hello World', link: '/docs/getting-started/hello-world' },
            { text: 'Key Concepts', link: '/docs/getting-started/key-concepts' },
          ]
        },
        {
          text: 'Language Guide',
          collapsed: false,
          items: [
            { text: 'Variables & Types', link: '/docs/language/variables-and-types' },
            { text: 'Functions', link: '/docs/language/functions' },
            { text: 'Control Flow', link: '/docs/language/control-flow' },
            { text: 'Structs & Enums', link: '/docs/language/structs-and-enums' },
            { text: 'Error Handling', link: '/docs/language/error-handling' },
            { text: 'Modules & Imports', link: '/docs/language/modules' },
            { text: 'Pattern Matching', link: '/docs/language/pattern-matching' },
            { text: 'Pipe Operator', link: '/docs/language/pipe-operator' },
            { text: 'Methods', link: '/docs/language/methods' },
            { text: 'Concurrency', link: '/docs/language/concurrency' },
          ]
        },
        {
          text: 'Agentic',
          collapsed: false,
          items: [
            { text: 'Providers', link: '/docs/agentic/providers' },
            { text: 'Tools', link: '/docs/agentic/tools' },
            { text: 'Agents', link: '/docs/agentic/agents' },
            { text: 'Workflows', link: '/docs/agentic/workflows' },
            { text: 'Streaming', link: '/docs/agentic/streaming' },
            { text: 'Agent Handoffs', link: '/docs/agentic/handoffs' },
            { text: 'Memory & Sessions', link: '/docs/agentic/memory' },
            { text: 'Generative UI', link: '/docs/agentic/generative-ui' },
            { text: 'ARP Protocol', link: '/docs/agentic/arp' },
          ]
        },
        {
          text: 'Standard Library',
          collapsed: true,
          items: [
            { text: 'Overview', link: '/docs/stdlib/overview' },
            { text: 'HTTP & Server', link: '/docs/stdlib/http' },
            { text: 'IO & File System', link: '/docs/stdlib/io' },
            { text: 'JSON', link: '/docs/stdlib/json' },
            { text: 'Postgres', link: '/docs/stdlib/postgres' },
            { text: 'Strings', link: '/docs/stdlib/strings' },
          ]
        },
        {
          text: 'Reference',
          collapsed: true,
          items: [
            { text: 'Examples', link: '/docs/examples' },
            { text: 'Grammar', link: '/docs/reference/grammar' },
            { text: 'Compiler Architecture', link: '/docs/reference/compiler' },
          ]
        },
      ]
    },

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
