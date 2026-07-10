<template>
  <div class="dashboard">
    <!-- 欢迎横幅 -->
    <section class="hero">
      <div class="hero-text">
        <h1 class="hero-title">调度仪表盘</h1>
        <p class="hero-sub">{{ greeting }}，{{ profile?.nickname || profile?.username || 'admin' }} · 调度中心实时运行数据一览</p>
      </div>
      <div class="hero-brand">
        <el-icon :size="56"><DataLine /></el-icon>
      </div>
    </section>

    <!-- 统计卡片 -->
    <section class="stats" v-loading="loading">
      <div
        v-for="card in cards"
        :key="card.label"
        class="stat-card"
        :style="{ '--accent': card.color }"
      >
        <div class="stat-icon">
          <el-icon :size="22"><component :is="card.icon" /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ card.value }}</div>
          <div class="stat-label">{{ card.label }}</div>
        </div>
      </div>
    </section>

    <!-- 快捷入口 -->
    <section class="quick">
      <div class="section-title">快捷入口</div>
      <div class="quick-grid">
        <router-link
          v-for="q in quicks"
          :key="q.to"
          :to="q.to"
          class="quick-card"
        >
          <el-icon :size="26"><component :is="q.icon" /></el-icon>
          <span class="quick-label">{{ q.label }}</span>
          <el-icon class="quick-arrow"><ArrowRight /></el-icon>
        </router-link>
      </div>
    </section>

    <!-- 说明 -->
    <section class="note">
      <div class="section-title">说明</div>
      <p class="note-text">
        任务总数与运行中任务反映调度规模；今日调度统计反映当日执行情况。前往
        <router-link to="/jobs" class="link">任务管理</router-link>·
        <router-link to="/executors" class="link">执行器管理</router-link>·
        <router-link to="/logs" class="link">调度日志</router-link>
        查看详情。
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/app/stores/user'
import { api } from '@/lib/http/client'
import {
  DataLine, List, VideoPlay, Connection, Calendar, CircleCheck, CircleClose,
  ArrowRight, Setting, Document, Tickets,
} from '@element-plus/icons-vue'

const userStore = useUserStore()
const profile = computed(() => userStore.profile)

const data = ref<any>({})
const loading = ref(false)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '凌晨好'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const cards = computed(() => [
  { label: '任务总数', value: data.value.job_total || 0, color: '#409eff', icon: List },
  { label: '运行中任务', value: data.value.job_running || 0, color: '#67c23a', icon: VideoPlay },
  { label: '执行器分组', value: data.value.exec_total || 0, color: '#909399', icon: Connection },
  { label: '今日调度', value: data.value.today_total || 0, color: '#e6a23c', icon: Calendar },
  { label: '今日成功', value: data.value.today_success || 0, color: '#67c23a', icon: CircleCheck },
  { label: '今日失败', value: data.value.today_fail || 0, color: '#f56c6c', icon: CircleClose },
])

const quicks = [
  { to: '/jobs', label: '任务管理', icon: Tickets },
  { to: '/executors', label: '执行器管理', icon: Setting },
  { to: '/logs', label: '调度日志', icon: Document },
]

async function load() {
  loading.value = true
  try {
    data.value = await api.get<any>('/api/dashboard')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.dashboard {
  padding: 16px 20px 24px;
}

/* 欢迎横幅 */
.hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  margin-bottom: 18px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-primary-light-3) 100%);
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.08);
}
.hero-title {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 600;
}
.hero-sub {
  margin: 0;
  font-size: 13px;
  opacity: 0.92;
}
.hero-brand {
  opacity: 0.85;
}

/* 统计卡片 */
.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 22px;
}
.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px 22px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  border-left: 4px solid var(--accent);
  transition: box-shadow 0.2s, transform 0.2s;
}
.stat-card:hover {
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 10px;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 12%, transparent);
}
.stat-value {
  font-size: 26px;
  font-weight: 600;
  line-height: 1.1;
  color: var(--el-text-color-primary);
  font-variant-numeric: tabular-nums;
}
.stat-label {
  margin-top: 4px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

/* 区块标题 */
.section-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

/* 快捷入口 */
.quick {
  margin-bottom: 22px;
}
.quick-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.quick-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 22px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  color: var(--el-text-color-regular);
  transition: box-shadow 0.2s, transform 0.2s, border-color 0.2s;
}
.quick-card:hover {
  border-color: var(--el-color-primary);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
  color: var(--el-color-primary);
}
.quick-label {
  flex: 1;
  font-size: 15px;
  font-weight: 500;
}
.quick-arrow {
  color: var(--el-text-color-placeholder);
}

/* 说明 */
.note-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.8;
  color: var(--el-text-color-regular);
}
.link {
  margin: 0 6px;
  color: var(--el-color-primary);
}

@media (max-width: 992px) {
  .stats,
  .quick-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 600px) {
  .stats,
  .quick-grid {
    grid-template-columns: 1fr;
  }
  .hero {
    flex-direction: column;
    text-align: center;
  }
}
</style>
