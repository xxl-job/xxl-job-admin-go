<template>
  <div class="p-4">
    <div class="mb-4 flex items-center gap-2">
      <h2 class="text-lg font-bold mr-4">调度日志</h2>
      <el-input v-model="jobId" placeholder="任务ID" style="width: 130px" clearable />
      <el-select v-model="status" placeholder="状态" style="width: 120px" clearable>
        <el-option value="success" label="成功" />
        <el-option value="fail" label="失败" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="px-4 py-2 text-sm">
            <div class="mb-2"><b>请求参数：</b>{{ row.executor_param || '(空)' }}</div>
            <div class="mb-2"><b>触发备注：</b>{{ row.trigger_msg || '(无)' }}</div>
            <div><b>执行消息：</b>{{ row.handle_msg || '(无)' }}</div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="id" label="日志ID" width="90" />
      <el-table-column prop="job_id" label="任务ID" width="90" />
      <el-table-column prop="executor_handler" label="Handler" width="160" />
      <el-table-column prop="executor_param" label="请求参数" min-width="160" show-overflow-tooltip />
      <el-table-column prop="executor_address" label="执行器" width="200" />
      <el-table-column label="触发时间" width="170">
        <template #default="{ row }">{{ fmt(row.trigger_time) }}</template>
      </el-table-column>
      <el-table-column label="结果" width="100">
        <template #default="{ row }">
          <el-tag :type="resultType(row)">{{ resultText(row) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-if="row.handle_code === 0 && row.trigger_code === 200" type="danger" @click="cancelRun(row)">终止</el-button>
          <el-button size="small" @click="openLogStream(row)">实时日志</el-button>
        </template>
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

    <el-dialog v-model="logDialogVisible" :title="`实时日志 #${currentLogId}`" width="800px" @close="closeLogStream">
      <div class="log-box" style="background:#1e1e1e;color:#d4d4d4;padding:12px;border-radius:4px;max-height:480px;overflow:auto;font-family:monospace;font-size:13px;white-space:pre-wrap;">
        <span v-if="!logContent" style="color:#888">等待日志输出…（执行器回传 LogChunk 后实时推送）</span>
        <span v-else>{{ logContent }}</span>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/lib/http/client'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const loading = ref(false)
const jobId = ref('')
const status = ref('')

const logDialogVisible = ref(false)
const currentLogId = ref(0)
const logContent = ref('')
let ws: WebSocket | null = null

function fmt(t: string): string {
  if (!t) return ''
  return new Date(t).toLocaleString()
}
function resultType(row: any): string {
  if (row.trigger_code !== 200) return 'danger'        // 触发失败
  if (row.handle_code === 0) return 'info'             // 运行中
  return row.handle_code === 200 ? 'success' : 'danger'
}
function resultText(row: any): string {
  if (row.trigger_code !== 200) return '触发失败'
  if (row.handle_code === 0) return '运行中'
  return row.handle_code === 200 ? '成功' : '失败'
}

async function cancelRun(row: any) {
  try {
    const res = await api.post<any>(`/api/logs/${row.id}/cancel`)
    ElMessage[res.canceled ? 'success' : 'info'](res.canceled ? '已下发终止' : '任务已结束')
    if (res.canceled) load()
  } catch (e: any) {
    ElMessage.error(e?.message || '终止失败')
  }
}

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/logs', {
      params: { page: page.value, size: size.value, job_id: jobId.value, status: status.value }
    })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function openLogStream(row: any) {
  currentLogId.value = row.id
  logContent.value = ''
  logDialogVisible.value = true
  const token = sessionStorage.getItem('va:access') || ''
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  ws = new WebSocket(`${proto}://${location.host}/api/logs/${row.id}/stream?token=${encodeURIComponent(token)}`)
  ws.onmessage = (e) => {
    logContent.value += e.data
  }
  ws.onerror = () => {
    logContent.value += '\n[连接错误]'
  }
}

function closeLogStream() {
  if (ws) {
    ws.close()
    ws = null
  }
  logContent.value = ''
}

onMounted(load)
</script>
