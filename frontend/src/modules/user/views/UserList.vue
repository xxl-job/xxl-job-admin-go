<template>
  <div class="p-4">
    <div class="mb-4 flex justify-between items-center">
      <div class="flex items-center gap-2">
        <h2 class="text-lg font-bold mr-4">用户管理</h2>
        <el-input v-model="filterKw" placeholder="用户名" style="width:200px" clearable />
      </div>
      <el-button type="primary" @click="openCreate">新建用户</el-button>
    </div>

    <el-table :data="filteredList" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="用户名" width="160" />
      <el-table-column label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="row.role === 1 ? 'danger' : 'info'">{{ row.role === 1 ? '管理员' : '普通' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="permission" label="执行器权限" min-width="220" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="row.id === 1" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑用户' : '新建用户'" width="520px">
      <el-form :model="editing" label-width="110px">
        <el-form-item label="用户名"><el-input v-model="editing.username" :disabled="!!editing.id" /></el-form-item>
        <el-form-item :label="editing.id ? '新密码' : '密码'">
          <el-input v-model="editing.password" type="password" show-password :placeholder="editing.id ? '留空不修改' : ''" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editing.role">
            <el-option :value="0" label="普通用户" />
            <el-option :value="1" label="管理员" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行器权限">
          <el-input v-model="editing.permission" placeholder="逗号分隔的 app_name，管理员免填" />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/lib/http/client'

const list = ref<any[]>([])
const loading = ref(false)
const filterKw = ref('')
const filteredList = computed(() => {
  if (!filterKw.value) return list.value
  const kw = filterKw.value.toLowerCase()
  return list.value.filter((it: any) => (it.username || '').toLowerCase().includes(kw))
})
const dialogVisible = ref(false)
const editing = reactive<any>({ id: undefined, username: '', password: '', role: 0, permission: '' })

async function load() {
  loading.value = true
  try {
    const res = await api.get<any>('/api/users')
    list.value = res.list || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(editing, { id: undefined, username: '', password: '', role: 0, permission: '' })
  dialogVisible.value = true
}

function openEdit(row: any) {
  Object.assign(editing, { id: row.id, username: row.username, password: '', role: row.role, permission: row.permission || '' })
  dialogVisible.value = true
}

async function save() {
  if (editing.id) {
    const body: any = { role: editing.role, permission: editing.permission }
    if (editing.password) body.password = editing.password
    await api.put(`/api/users/${editing.id}`, body)
  } else {
    await api.post('/api/users', editing)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  load()
}

async function remove(row: any) {
  await ElMessageBox.confirm('确认删除该用户?', '提示', { type: 'warning' })
  await api.del(`/api/users/${row.id}`)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>
