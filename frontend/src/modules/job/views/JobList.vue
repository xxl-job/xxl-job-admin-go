<template>
  <div class="p-4">
    <div class="mb-4 flex justify-between items-center">
      <div class="flex items-center gap-2">
        <h2 class="text-lg font-bold mr-4">任务管理</h2>
        <el-input v-model="searchKw" placeholder="描述/Handler" style="width:200px" clearable @keyup.enter="onSearch" />
        <el-select v-model="searchStatus" placeholder="状态" style="width:120px" clearable>
          <el-option value="running" label="运行中" />
          <el-option value="stopped" label="已停止" />
        </el-select>
        <el-button type="primary" @click="onSearch">查询</el-button>
        <el-button @click="onReset">重置</el-button>
      </div>
      <el-button type="primary" @click="openCreate">新建任务</el-button>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="job_desc" label="描述" min-width="140" />
      <el-table-column prop="executor_handler" label="Handler" width="160" />
      <el-table-column label="执行器" width="140">
        <template #default="{ row }">{{ executorName(row.job_group) }}</template>
      </el-table-column>
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
          <el-button size="small" @click="openTrigger(row)">触发</el-button>
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

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑任务' : '新建任务'" width="780px">
      <el-form :model="editing" label-width="120px">
        <el-collapse v-model="activeSections">
          <el-collapse-item title="基础信息" name="basic">
            <el-row :gutter="16">
              <el-col :span="12"><el-form-item label="描述"><el-input v-model="editing.job_desc" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="执行器分组">
                <el-select v-model="editing.job_group" placeholder="选择执行器分组" filterable style="width:100%">
                  <el-option v-for="ex in executorOptions" :key="ex.id" :value="ex.id"
                    :label="`${ex.app_name}${ex.title ? ' (' + ex.title + ')' : ''}`" />
                </el-select>
              </el-form-item></el-col>
              <el-col :span="12"><el-form-item label="Handler"><el-input v-model="editing.executor_handler" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="参数"><el-input v-model="editing.executor_param" /></el-form-item></el-col>
            </el-row>
          </el-collapse-item>

          <el-collapse-item title="调度配置" name="schedule">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="调度类型">
                  <el-select v-model="editing.schedule_type">
                    <el-option :value="1" label="CRON" />
                    <el-option :value="2" label="固定频率(秒)" />
                    <el-option :value="3" label="固定延时(秒)" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12"><el-form-item label="调度配置"><el-input v-model="editing.schedule_conf" placeholder="CRON 表达式或秒数" /></el-form-item></el-col>
            </el-row>
            <div v-if="previewTimes.length" class="preview-box">
              <span class="text-gray-500">接下来 5 次触发：</span>
              <span v-for="(t, i) in previewTimes" :key="i" class="preview-item">{{ fmt(t) }}</span>
            </div>
            <div v-if="previewError" class="text-red-500 text-sm">{{ previewError }}</div>
            <el-row :gutter="16" class="mt-2">
              <el-col :span="12">
                <el-form-item label="GLUE类型">
                  <el-select v-model="editing.glue_type">
                    <el-option :value="0" label="BEAN（已注册 handler）" />
                    <el-option :value="1" label="GLUE（Go 在线代码）" />
                    <el-option :value="2" label="Shell 脚本" />
                    <el-option :value="3" label="Python 脚本" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item v-if="editing.glue_type !== 0" label="GLUE 代码">
              <el-input v-model="editing.glue_source" type="textarea" :rows="8" placeholder='func Run(param string) (int, string) {&#10;  return 200, "ok"&#10;}' />
            </el-form-item>
          </el-collapse-item>

          <el-collapse-item title="路由与阻塞" name="route">
            <el-row :gutter="16">
              <el-col :span="12">
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
              </el-col>
              <el-col :span="12">
                <el-form-item label="阻塞策略">
                  <el-select v-model="editing.executor_block_strategy">
                    <el-option :value="0" label="串行" />
                    <el-option :value="1" label="丢弃后续" />
                    <el-option :value="2" label="覆盖之前" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12"><el-form-item label="超时(秒)"><el-input-number v-model="editing.executor_timeout" :min="0" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="失败重试"><el-input-number v-model="editing.executor_fail_retry_count" :min="0" /></el-form-item></el-col>
            </el-row>
          </el-collapse-item>

          <el-collapse-item title="高级" name="advanced">
            <el-row :gutter="16">
              <el-col :span="12"><el-form-item label="子任务ID"><el-input v-model="editing.child_jobid" placeholder="逗号分隔的子任务ID，父任务成功后触发" /></el-form-item></el-col>
              <el-col :span="12"><el-form-item label="流量泳道"><el-input v-model="editing.cluster" placeholder="留空=不限" /></el-form-item></el-col>
            </el-row>
          </el-collapse-item>
        </el-collapse>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveJob">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="trigDialog" title="手动执行一次" width="480px">
      <el-form label-width="100px">
        <el-form-item label="任务">
          <span>#{{ trigCtx.id }} [{{ trigCtx.executor_handler }}]</span>
        </el-form-item>
        <el-form-item label="临时参数">
          <el-input v-model="trigParam" placeholder="留空则用任务默认参数" />
        </el-form-item>
        <el-form-item label="指定机器">
          <el-select v-model="trigAddr" placeholder="自动路由" clearable :disabled="!onlineAddrs.length">
            <el-option v-for="a in onlineAddrs" :key="a" :value="a" :label="a" />
          </el-select>
          <div v-if="!onlineAddrs.length" class="text-gray-400 text-xs">该分组暂无在线执行器</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="trigDialog = false">取消</el-button>
        <el-button type="primary" @click="doTrigger">立即执行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
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
  cluster: string
  trigger_status?: number
}

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const loading = ref(false)
// 执行器分组选项（任务表单下拉 + 列表名称映射用）
const executorOptions = ref<any[]>([])

async function loadExecutors() {
  try {
    const res = await api.get<any>('/api/executors')
    executorOptions.value = (res.list || []).map((it: any) => it.executor).filter((e: any) => e && e.id !== 0)
  } catch {
    executorOptions.value = []
  }
}
function executorName(jobGroup: number): string {
  const ex = executorOptions.value.find((e: any) => e.id === jobGroup)
  return ex ? ex.app_name : `#${jobGroup}`
}
const dialogVisible = ref(false)
const editing = reactive<Job>(emptyJob())

// 表单分区折叠
const activeSections = ref(['basic', 'schedule', 'route', 'advanced'])

// CRON 预览
const previewTimes = ref<number[]>([])
const previewError = ref('')
let previewTimer: any = null

watch(() => [editing.schedule_type, editing.schedule_conf], () => {
  if (previewTimer) clearTimeout(previewTimer)
  previewTimer = setTimeout(loadPreview, 300)
}, { deep: true })

async function loadPreview() {
  previewError.value = ''
  if (!editing.schedule_conf) { previewTimes.value = []; return }
  try {
    const res = await api.get<any>('/api/jobs/preview-schedule', {
      params: { schedule_type: editing.schedule_type, schedule_conf: editing.schedule_conf }
    })
    previewTimes.value = res.next_times || []
  } catch (e: any) {
    previewTimes.value = []
    previewError.value = e?.message || '解析失败'
  }
}

function fmt(ms: number): string {
  return new Date(ms).toLocaleString()
}

// 手动触发弹窗
const trigDialog = ref(false)
const trigCtx = reactive<any>({ id: 0, executor_handler: '', job_group: 0 })
const trigParam = ref('')
const trigAddr = ref('')
const onlineAddrs = ref<string[]>([])

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
    glue_source: '',
    cluster: ''
  }
}

function scheduleText(row: any): string {
  const types: Record<number, string> = { 1: 'CRON', 2: '频率', 3: '延时' }
  return `${types[row.schedule_type] || '?'}: ${row.schedule_conf}`
}

// 搜索
const searchKw = ref('')
const searchStatus = ref('')

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/jobs', { params: { page: page.value, size: size.value, keyword: searchKw.value, status: searchStatus.value } })
    list.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function onSearch() { page.value = 1; load() }
function onReset() { searchKw.value = ''; searchStatus.value = ''; page.value = 1; load() }

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
    glue_source: row.glue_source || '',
    cluster: row.cluster || ''
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

async function openTrigger(row: any) {
  Object.assign(trigCtx, { id: row.id, executor_handler: row.executor_handler, job_group: row.job_group })
  trigParam.value = ''
  trigAddr.value = ''
  // 取该分组在线地址
  try {
    const res = await api.get<any>('/api/executors')
    const items = res.list || []
    // 优先按 job_group 匹配 executor.id，其次匹配 app_name
    let hit = items.find((it: any) => it.executor?.id === row.job_group)
    if (!hit) hit = items.find((it: any) => it.executor?.app_name)
    onlineAddrs.value = hit?.online || []
  } catch {
    onlineAddrs.value = []
  }
  trigDialog.value = true
}

async function doTrigger() {
  try {
    await api.post(`/api/jobs/${trigCtx.id}/trigger`, {
      executor_param: trigParam.value,
      executor_address: trigAddr.value
    })
    trigDialog.value = false
    ElMessageBox.confirm('已触发，是否立即查看调度日志？', '提示', { type: 'success' })
      .then(() => { window.location.hash = '#/logs?job_id=' + trigCtx.id })
      .catch(() => {})
  } catch (e: any) {
    ElMessage.error(e?.message || '触发失败')
  }
}

async function removeJob(row: any) {
  await ElMessageBox.confirm('确认删除该任务?', '提示', { type: 'warning' })
  await api.del(`/api/jobs/${row.id}`)
  ElMessage.success('已删除')
  load()
}

onMounted(() => {
  load()
  loadExecutors()
})
</script>

<style scoped>
.preview-box {
  margin-top: 8px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
}
.preview-item {
  display: inline-block;
  margin: 2px 8px 2px 0;
  color: #409eff;
}
</style>
