<template>
  <SearchTable
    :title="t('dept.title')"
    :loading="loading"
    :data="tableData"
    :columns="columns"
    :pagination="pagination"
    :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
    default-expand-all
  >
    <template #search>
      <el-input
        v-model="searchForm.keyword"
        :placeholder="t('dept.searchKeyword')"
        clearable
        style="width: 200px"
        @keyup.enter="fetchList"
      />
      <el-select
        v-model="searchForm.status"
        clearable
        :placeholder="t('common.column.status')"
        style="width: 120px"
      >
        <el-option
          :label="t('common.status.enable')"
          value="active"
        />
        <el-option
          :label="t('common.status.disable')"
          value="inactive"
        />
      </el-select>
    </template>

    <template #actions>
      <el-button
        type="primary"
        :icon="Plus"
        @click="openDrawer('add')"
      >
        {{ t('common.action.create') }}
      </el-button>
      <el-button
        :icon="Refresh"
        @click="fetchList"
      >
        {{ t('common.action.refresh') }}
      </el-button>
    </template>

    <template #col-status="{ row }">
      <el-tag
        :type="row.status === 'active' ? 'success' : 'danger'"
        size="small"
      >
        {{ row.status === 'active' ? t('common.status.enable') : t('common.status.disable') }}
      </el-tag>
    </template>

    <template #col-createTime="{ row }">
      {{ formatDate(row.createTime) }}
    </template>

    <template #col-actions="{ row }">
      <el-button
        link
        type="primary"
        size="small"
        @click="openDrawer('add', row)"
      >
        {{ t('dept.addSubLabel') }}
      </el-button>
      <el-button
        link
        type="primary"
        size="small"
        @click="openDrawer('edit', row)"
      >
        {{ t('common.action.edit') }}
      </el-button>
      <el-button
        link
        type="danger"
        size="small"
        @click="handleDelete(row.id)"
      >
        {{ t('common.action.delete') }}
      </el-button>
    </template>
  </SearchTable>

  <DeptFormDrawer
    v-model="drawerVisible"
    :mode="drawerMode"
    :data="editingRow"
    @success="onFormSuccess"
  />
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { confirmService } from '@/lib/confirm'
import { SearchTable } from '@/app/components'
import { formatDate } from '@/lib/format'
import { t } from '@/lib/i18n'
import {
  fetchDeptTree,
  deleteDept,
  type DeptInfo,
  type DeptSearchRequest,
} from '../api'
import DeptFormDrawer from './DeptFormDrawer.vue'

const loading = ref(false)
const tableData = ref<DeptInfo[]>([])
const pagination = ref({ page: 1, size: 10, total: 0 })

const searchForm = reactive<DeptSearchRequest>({
  keyword: '',
  status: '',
})

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit'>('add')
const editingRow = ref<DeptInfo | null>(null)

const fetchList = async () => {
  loading.value = true
  try {
    const data = await fetchDeptTree(searchForm)
    tableData.value = data
  } finally {
    loading.value = false
  }
}

const openDrawer = (mode: 'add' | 'edit', row?: DeptInfo) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}

const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}

const handleDelete = async (id: string) => {
  const confirmed = await confirmService.showConfirm(t('dept.deleteConfirm'))
  if (!confirmed) return
  try {
    await deleteDept(id)
    ElMessage.success(t('common.message.deleteSuccess'))
    fetchList()
  } catch {
    // 失败由 http 拦截器提示
  }
}

const columns = computed(() => [
  { prop: 'name', label: t('dept.field.name'), minWidth: 200 },
  { prop: 'leader', label: t('dept.field.leader'), minWidth: 120 },
  { prop: 'phone', label: t('dept.field.phone'), minWidth: 150 },
  { prop: 'status', label: t('common.column.status'), minWidth: 100, slot: 'status' },
  { prop: 'sort', label: t('dept.field.sort'), minWidth: 100 },
  { prop: 'createTime', label: t('common.column.createTime'), minWidth: 180, slot: 'createTime' },
  { prop: 'actions', label: t('common.column.actions'), minWidth: 250, slot: 'actions' },
])

onMounted(fetchList)
</script>
