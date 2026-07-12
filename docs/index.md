---
layout: home

hero:
  name: XXL-JOB-Admin-Go
  text: Go 实现的分布式任务调度中心
  tagline: XXL-JOB 调度中心的 Go 现代化重实现 · gRPC 长连接调度 · MySQL 租约锁 · 无 Redis
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/getting-started
    - theme: alt
      text: 镜像部署
      link: /guide/docker-images

features:
  - title: gRPC 调度
    details: 非兼容 Java XXL-JOB，gRPC ExecutorLink 通道连接执行器
  - title: 无 Redis 依赖
    details: MySQL 租约锁选主，单二进制 + MySQL 即可部署
  - title: vue-admin 前端
    details: Vue 3 + Element Plus 管理界面，开箱即用
  - title: Connect 协议
    details: 前端 Connect-Web 直连，实时日志 WebSocket 流
---
