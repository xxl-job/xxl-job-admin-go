<template>
  <div class="p-4">
    <div class="mb-4 flex justify-between items-center">
      <h2 class="text-lg font-bold">执行器管理</h2>
      <div>
        <el-button type="primary" @click="openCreate">新建执行器</el-button>
        <el-button @click="load" :loading="loading">刷新</el-button>
      </div>
    </div>

    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="executor.id" label="ID" width="70" />
      <el-table-column prop="executor.app_name" label="AppName" width="160" />
      <el-table-column prop="executor.title" label="标题" min-width="140">
        <template #default="{ row }">
          {{ row.executor.title }}
          <el-tag v-if="row.executor.id === 0" size="small" type="info" class="ml-1">未配置分组</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册方式" width="110">
        <template #default="{ row }">
          <el-tag :type="row.executor.address_type === 0 ? 'success' : 'warning'">
            {{ row.executor.address_type === 0 ? '自动注册' : '手动录入' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="在线机器" min-width="220">
        <template #default="{ row }">
          <el-badge :value="row.online?.length || 0" type="primary" class="mr-2">
            <el-tag size="small">实例</el-tag>
          </el-badge>
          <el-tag v-for="addr in row.online" :key="addr" size="small" class="mr-1 mb-1" type="success"
            style="cursor: pointer" :title="`点击复制 ${addr}`" @click="copyAddr(addr)">
            <span class="dot-online">●</span>{{ fmtAddr(addr) }}
          </el-tag>
          <span v-if="!row.online || row.online.length === 0" class="text-gray-400">无在线机器</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" :disabled="row.executor.id === 0" @click="openEdit(row.executor)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="row.executor.id === 0" @click="remove(row.executor)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑执行器' : '新建执行器'" width="540px">
      <el-form :model="editing" label-width="110px">
        <el-form-item label="AppName"><el-input v-model="editing.app_name" :disabled="!!editing.id" /></el-form-item>
        <el-form-item label="标题"><el-input v-model="editing.title" /></el-form-item>
        <el-form-item label="注册方式">
          <el-select v-model="editing.address_type">
            <el-option :value="0" label="自动注册" />
            <el-option :value="1" label="手动录入" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="editing.address_type === 1" label="地址列表">
          <el-input v-model="editing.address_list" type="textarea" :rows="3" placeholder="多个地址逗号分隔，如 http://1.1.1.1:9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/lib/http/client'

const list = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editing = reactive<any>({ id: undefined, app_name: '', title: '', address_type: 0, address_list: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/executors')
    list.value = res.list || []
  } finally {
    loading.value = false
  }
}

// fmtAddr 地址友好显示：[::1]:59659 -> localhost:59659；[ipv6]:port -> ipv6:port
function fmtAddr(addr: string): string {
  return addr.replace(/^\[::1\]/, 'localhost').replace(/^\[(.+)\]:(\d+)$/, '$1:$2')
}

// copyAddr 点击复制完整地址
async function copyAddr(addr: string) {
  try {
    await navigator.clipboard.writeText(addr)
    ElMessage.success('已复制: ' + addr)
  } catch {
    ElMessage.warning('复制失败，请手动选择')
  }
}

function openCreate() {
  Object.assign(editing, { id: undefined, app_name: '', title: '', address_type: 0, address_list: '' })
  dialogVisible.value = true
}

function openEdit(row: any) {
  Object.assign(editing, {
    id: row.id,
    app_name: row.app_name,
    title: row.title,
    address_type: row.address_type,
    address_list: row.address_list || ''
  })
  dialogVisible.value = true
}

async function save() {
  if (editing.id) {
    await api.put(`/api/executors/${editing.id}`, editing)
  } else {
    await api.post('/api/executors', editing)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  load()
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该执行器分组?', '提示', { type: 'warning' })
  await api.del(`/api/executors/${row.id}`)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<style scoped>
.dot-online { color: #67c23a; margin-right: 2px; }
</style>
