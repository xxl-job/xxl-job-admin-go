<template>
  <div class="p-4">
    <div class="mb-4 flex items-center gap-2">
      <h2 class="text-lg font-bold mr-4">审计日志</h2>
      <el-input v-model="searchActor" placeholder="操作人" style="width:160px" clearable @clear="onSearch" @keyup.enter="onSearch" />
      <el-input v-model="searchAction" placeholder="操作" style="width:180px" clearable @clear="onSearch" @keyup.enter="onSearch" />
      <el-button type="primary" :loading="loading" @click="onSearch">查询</el-button>
      <el-button @click="onReset">重置</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="actor" label="操作人" width="120" />
      <el-table-column prop="action" label="操作" width="220" show-overflow-tooltip />
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
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      @current-change="load"
      @size-change="onSizeChange"
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

const searchActor = ref('')
const searchAction = ref('')

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/audit', { params: { page: page.value, size: size.value, actor: searchActor.value, action: searchAction.value } })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function onSearch() { page.value = 1; load() }
function onSizeChange() { page.value = 1; load() }
function onReset() { searchActor.value = ''; searchAction.value = ''; page.value = 1; load() }

onMounted(load)
</script>
