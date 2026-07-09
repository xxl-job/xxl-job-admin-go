<template>
  <div class="p-4">
    <div class="mb-4 flex justify-between items-center">
      <h2 class="text-lg font-bold">任务管理</h2>
      <el-button type="primary" @click="openCreate">新建任务</el-button>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="job_desc" label="描述" min-width="140" />
      <el-table-column prop="executor_handler" label="Handler" width="160" />
      <el-table-column label="调度" width="180">
        <template #default="{ row }">
          {{ scheduleText(row) }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.trigger_status === 1 ? 'success' : 'info'">
            {{ row.trigger_status === 1 ? '运行' : '停止' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="340" fixed="right">
        <template #default="{ row }">
          <el-button size="small" v-if="row.trigger_status !== 1" type="success" @click="startJob(row)">启动</el-button>
          <el-button size="small" v-else type="warning" @click="stopJob(row)">停止</el-button>
          <el-button size="small" @click="triggerJob(row)">触发</el-button>
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="removeJob(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑任务' : '新建任务'" width="640px">
      <el-form :model="editing" label-width="120px">
        <el-form-item label="描述"><el-input v-model="editing.job_desc" /></el-form-item>
        <el-form-item label="执行器分组ID"><el-input-number v-model="editing.job_group" :min="1" /></el-form-item>
        <el-form-item label="调度类型">
          <el-select v-model="editing.schedule_type">
            <el-option :value="1" label="CRON" />
            <el-option :value="2" label="固定频率(秒)" />
            <el-option :value="3" label="固定延时(秒)" />
          </el-select>
        </el-form-item>
        <el-form-item label="调度配置"><el-input v-model="editing.schedule_conf" placeholder="CRON 表达式或秒数" /></el-form-item>
        <el-form-item label="Handler"><el-input v-model="editing.executor_handler" /></el-form-item>
        <el-form-item label="参数"><el-input v-model="editing.executor_param" /></el-form-item>
        <el-form-item label="路由策略">
          <el-select v-model="editing.executor_route_strategy">
            <el-option :value="1" label="第一个" />
            <el-option :value="2" label="最后一个" />
            <el-option :value="3" label="轮询" />
            <el-option :value="4" label="随机" />
            <el-option :value="5" label="一致性HASH" />
            <el-option :value="10" label="分片广播" />
          </el-select>
        </el-form-item>
        <el-form-item label="阻塞策略">
          <el-select v-model="editing.executor_block_strategy">
            <el-option :value="0" label="串行" />
            <el-option :value="1" label="丢弃后续" />
            <el-option :value="2" label="覆盖之前" />
          </el-select>
        </el-form-item>
        <el-form-item label="超时(秒)"><el-input-number v-model="editing.executor_timeout" :min="0" /></el-form-item>
        <el-form-item label="失败重试"><el-input-number v-model="editing.executor_fail_retry_count" :min="0" /></el-form-item>
        <el-form-item label="子任务ID">
          <el-input v-model="editing.child_jobid" placeholder="逗号分隔的子任务ID，父任务成功后触发" />
        </el-form-item>
        <el-form-item label="GLUE类型">
          <el-select v-model="editing.glue_type">
            <el-option :value="0" label="BEAN（已注册 handler）" />
            <el-option :value="1" label="GLUE（Go 在线代码）" />
            <el-option :value="2" label="Shell 脚本" />
            <el-option :value="3" label="Python 脚本" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editing.glue_type !== 0" label="GLUE 代码">
          <el-input v-model="editing.glue_source" type="textarea" :rows="8" placeholder='func Run(param string) (int, string) {&#10;  return 200, "ok"&#10;}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveJob">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/lib/http/client'

interface Job {
  id?: number
  job_desc: string
  job_group: number
  schedule_type: number
  schedule_conf: string
  executor_handler: string
  executor_param: string
  executor_route_strategy: number
  executor_block_strategy: number
  executor_timeout: number
  executor_fail_retry_count: number
  child_jobid: string
  glue_type: number
  glue_source: string
  trigger_status?: number
}

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const loading = ref(false)
const dialogVisible = ref(false)
const editing = reactive<Job>(emptyJob())

function emptyJob(): Job {
  return {
    id: undefined,
    job_desc: '',
    job_group: 1,
    schedule_type: 1,
    schedule_conf: '0 * * * * ?',
    executor_handler: '',
    executor_param: '',
    executor_route_strategy: 1,
    executor_block_strategy: 0,
    executor_timeout: 0,
    executor_fail_retry_count: 0,
    child_jobid: '',
    glue_type: 0,
    glue_source: ''
  }
}

function scheduleText(row: any): string {
  const types: Record<number, string> = { 1: 'CRON', 2: '频率', 3: '延时' }
  return `${types[row.schedule_type] || '?'}: ${row.schedule_conf}`
}

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/jobs', { params: { page: page.value, size: size.value } })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(editing, emptyJob())
  dialogVisible.value = true
}

function openEdit(row: any) {
  Object.assign(editing, {
    id: row.id,
    job_desc: row.job_desc,
    job_group: row.job_group,
    schedule_type: row.schedule_type,
    schedule_conf: row.schedule_conf,
    executor_handler: row.executor_handler,
    executor_param: row.executor_param,
    executor_route_strategy: row.executor_route_strategy,
    executor_block_strategy: row.executor_block_strategy,
    executor_timeout: row.executor_timeout,
    executor_fail_retry_count: row.executor_fail_retry_count,
    child_jobid: row.child_jobid || '',
    glue_type: row.glue_type || 0,
    glue_source: row.glue_source || ''
  })
  dialogVisible.value = true
}

async function saveJob() {
  if (editing.id) {
    await api.put(`/api/jobs/${editing.id}`, editing)
  } else {
    await api.post('/api/jobs', editing)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  load()
}

async function startJob(row: any) {
  await api.post(`/api/jobs/${row.id}/start`)
  ElMessage.success('已启动')
  load()
}

async function stopJob(row: any) {
  await api.post(`/api/jobs/${row.id}/stop`)
  ElMessage.success('已停止')
  load()
}

async function triggerJob(row: any) {
  await api.post(`/api/jobs/${row.id}/trigger`)
  ElMessage.success('已触发')
}

async function removeJob(row: any) {
  await ElMessageBox.confirm('确认删除该任务?', '提示', { type: 'warning' })
  await api.del(`/api/jobs/${row.id}`)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>
