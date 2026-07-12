<template>
  <div class="p-4">
    <div class="mb-4 flex items-center gap-2 flex-wrap">
      <h2 class="text-lg font-bold mr-4">调度日志</h2>
      <el-input v-model="jobId" placeholder="任务ID" style="width: 130px" clearable @clear="onSearch" @keyup.enter="onSearch" />
      <el-input v-model="searchHandler" placeholder="Handler" style="width: 160px" clearable @clear="onSearch" @keyup.enter="onSearch" />
      <el-select v-model="status" placeholder="状态" style="width: 120px" clearable @change="onSearch">
        <el-option value="success" label="成功" />
        <el-option value="fail" label="失败" />
      </el-select>
      <el-date-picker v-model="timeRange" type="datetimerange" range-separator="至"
        start-placeholder="开始时间" end-placeholder="结束时间"
        value-format="YYYY-MM-DD HH:mm:ss" style="width:360px" @change="onSearch" />
      <el-button type="primary" :loading="loading" @click="onSearch">查询</el-button>
      <el-button @click="onReset">重置</el-button>
    </div>

    <el-table :data="list" v-loading="loading" border stripe empty-text="暂无调度日志，请调整筛选条件">
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
      <el-table-column label="任务" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.job_desc ? `${row.job_desc} (#${row.job_id})` : `#${row.job_id}` }}
        </template>
      </el-table-column>
      <el-table-column prop="executor_handler" label="Handler" width="160" show-overflow-tooltip />
      <el-table-column prop="executor_param" label="请求参数" min-width="160" show-overflow-tooltip />
      <el-table-column label="执行器" width="200">
        <template #default="{ row }">
          <div class="leading-tight">
            <span>{{ row.executor_app || '-' }}</span>
            <div v-if="row.executor_address" class="text-xs text-gray-400">{{ row.executor_address }}</div>
          </div>
        </template>
      </el-table-column>
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
          <el-button size="small" v-if="row.handle_code === 0 && row.trigger_code === 200" type="danger" :loading="cancellingId === row.id" @click="cancelRun(row)">终止</el-button>
          <el-button size="small" @click="openLogStream(row)">实时日志</el-button>
        </template>
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
const cancellingId = ref<number | null>(null)
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
  cancellingId.value = row.id
  try {
    const res = await api.post<any>(`/api/logs/${row.id}/cancel`)
    ElMessage[res.canceled ? 'success' : 'info'](res.canceled ? '已下发终止' : '任务已结束')
    if (res.canceled) load()
  } catch (e: any) {
    ElMessage.error(e?.message || '终止失败')
  } finally {
    cancellingId.value = null
  }
}

const searchHandler = ref('')
const timeRange = ref<string[]>([])

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value, job_id: jobId.value, status: status.value, handler: searchHandler.value }
    if (timeRange.value && timeRange.value.length === 2) {
      params.start_time = timeRange.value[0]
      params.end_time = timeRange.value[1]
    }
    const res = await api.get<any>('/api/logs', { params })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function onSearch() { page.value = 1; load() }
function onSizeChange() { page.value = 1; load() }
function onReset() {
  jobId.value = ''; searchHandler.value = ''; status.value = ''; timeRange.value = []
  page.value = 1; load()
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
