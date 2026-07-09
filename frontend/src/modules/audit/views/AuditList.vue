<template>
  <div class="p-4">
    <h2 class="text-lg font-bold mb-4">审计日志</h2>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="actor" label="操作人" width="120" />
      <el-table-column prop="action" label="操作" width="220" />
      <el-table-column prop="target" label="目标" min-width="200" show-overflow-tooltip />
      <el-table-column label="时间" width="180">
        <template #default="{ row }">{{ fmt(row.at) }}</template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="mt-4"
      v-model:current-page="page"
      v-model:page-size="size"
      :total="total"
      layout="total, prev, pager, next"
      @current-change="load"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/lib/http/client'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const loading = ref(false)

function fmt(t: string): string {
  return t ? new Date(t).toLocaleString() : ''
}

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/audit', { params: { page: page.value, size: size.value } })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
