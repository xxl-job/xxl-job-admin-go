import type { RouteRecordRaw } from 'vue-router'

// 静态路由：xxl-job 后端无 /api/system/menus 接口，菜单完全静态化。
// 仅保留业务路由 + 末尾 catchAll 兜底（替代已删的 about/NotFound.vue）。
const menus: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/modules/dashboard/views/Home.vue'),
    meta: {
      title: '首页',
      icon: 'HomeFilled',
      showMenu: true
    }
  },
  {
    path: '/jobs',
    name: 'jobs',
    component: () => import('@/modules/job/views/JobList.vue'),
    meta: {
      title: '任务管理',
      icon: 'Operation',
      showMenu: true
    }
  },
  {
    path: '/executors',
    name: 'executors',
    component: () => import('@/modules/executor/views/ExecutorList.vue'),
    meta: {
      title: '执行器管理',
      icon: 'Cpu',
      showMenu: true
    }
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('@/modules/log/views/LogList.vue'),
    meta: {
      title: '调度日志',
      icon: 'Tickets',
      showMenu: true
    }
  },
  {
    path: '/audit',
    name: 'audit',
    component: () => import('@/modules/audit/views/AuditList.vue'),
    meta: {
      title: '审计日志',
      icon: 'List',
      showMenu: true
    }
  },
  {
    path: '/users',
    name: 'users',
    component: () => import('@/modules/user/views/UserList.vue'),
    meta: {
      title: '用户管理',
      icon: 'UserFilled',
      showMenu: true
    }
  },
  // catchAll：未匹配路由重定向回首页（替代已删的 /404 -> NotFound.vue）
  {
    path: '/:catchAll(.*)*',
    name: 'catchAll',
    redirect: '/',
    meta: {}
  }
]

export default menus
