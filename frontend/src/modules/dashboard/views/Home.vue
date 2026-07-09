<template>
  <div class="p-4">
    <h2 class="text-lg font-bold mb-4">调度仪表盘</h2>
    <el-row :gutter="16" v-loading="loading">
      <el-col :span="8" v-for="card in cards" :key="card.label" class="mb-4">
        <el-card shadow="hover">
          <div class="text-gray-500 text-sm">{{ card.label }}</div>
          <div class="text-3xl font-bold mt-2" :style="{ color: card.color }">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="mt-2">
      <template #header><span class="font-bold">说明</span></template>
      <p class="text-gray-600 text-sm">
        本仪表盘展示调度中心实时运行数据。任务总数与运行中任务反映调度规模；今日调度统计反映当日执行情况。
        可通过 <router-link to="/jobs" class="text-blue-500">任务管理</router-link> 管理调度任务，
        <router-link to="/executors" class="text-blue-500">执行器管理</router-link> 查看在线机器，
        <router-link to="/logs" class="text-blue-500">调度日志</router-link> 排查执行结果。
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/lib/http/client'

const data = ref<any>({})
const loading = ref(false)

const cards = computed(() => [
  { label: '任务总数', value: data.value.job_total || 0, color: '#409eff' },
  { label: '运行中任务', value: data.value.job_running || 0, color: '#67c23a' },
  { label: '执行器分组', value: data.value.exec_total || 0, color: '#909399' },
  { label: '今日调度', value: data.value.today_total || 0, color: '#e6a23c' },
  { label: '今日成功', value: data.value.today_success || 0, color: '#67c23a' },
  { label: '今日失败', value: data.value.today_fail || 0, color: '#f56c6c' }
])

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
