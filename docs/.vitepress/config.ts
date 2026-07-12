import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'XXL-JOB-Admin-Go',
  description: 'Go 实现的分布式任务调度中心 · 用户使用文档',
  lang: 'zh-CN',
  lastUpdated: true,
  // 关闭 cleanUrls：生成 /xxx.html，nginx 无需 SPA fallback 即可访问
  cleanUrls: false,
  // 文档 base 可经 DOCS_BASE 覆盖：独立域名用 /，子路径用 /docs/
  base: process.env.DOCS_BASE ?? '/',
  ignoreDeadLinks: true,
  head: [['meta', { name: 'theme-color', content: '#409EFF' }]],
  themeConfig: {
    logo: '/logo.png',
    search: { provider: 'local' },
    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/guide/getting-started' },
      { text: '运维', link: '/operations' },
      { text: 'GitHub', link: 'https://github.com/xxl-job/xxl-job-admin-go' }
    ],
    sidebar: [
      {
        text: '入门',
        items: [
          { text: '快速开始', link: '/guide/getting-started' },
          { text: '仪表盘', link: '/guide/dashboard' }
        ]
      },
      {
        text: '核心功能',
        items: [
          { text: '执行器管理', link: '/guide/executor' },
          { text: '任务管理', link: '/guide/job' },
          { text: '调度日志', link: '/guide/log' }
        ]
      },
      {
        text: '部署',
        items: [
          { text: '镜像部署', link: '/guide/docker-images' }
        ]
      },
      {
        text: '系统管理',
        items: [{ text: '用户与权限', link: '/guide/user' }]
      },
      {
        text: '运维',
        items: [{ text: '运维与可观测', link: '/operations' }]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/xxl-job/xxl-job-admin-go' }
    ],
    footer: {
      message: '基于 MIT 协议发布',
      copyright: 'Copyright © 2026 XXL-JOB-Admin-Go'
    },
    outline: { level: [2, 3] },
    docFooter: { prev: '上一页', next: '下一页' },
    lastUpdatedText: '最后更新'
  }
})
