<template>
  <SearchTable
    :title="t('permission.title')"
    :loading="loading"
    :data="tableData"
    :columns="columns"
    :pagination="pagination"
    :selected-rows="tableSelectedRows"
    selectable
    row-key="id"
    @search="handleSearch"
    @reset="handleReset"
    @page-change="handlePageChange"
    @selection-change="onSelectionChange"
  >
    <template #search>
      <el-input
        v-model="searchForm.keyword"
        :placeholder="t('permission.searchKeyword')"
        clearable
        style="width: 220px"
        @keyup.enter="handleSearch"
      />
      <el-select
        v-model="searchForm.module"
        clearable
        :placeholder="t('permission.field.module')"
        style="width: 140px"
      >
        <el-option
          :label="t('permission.option.moduleSystem')"
          value="system"
        />
        <el-option
          :label="t('permission.option.moduleUser')"
          value="user"
        />
        <el-option
          :label="t('permission.option.moduleRole')"
          value="role"
        />
        <el-option
          :label="t('permission.option.modulePermission')"
          value="permission"
        />
        <el-option
          :label="t('permission.option.moduleDict')"
          value="dict"
        />
        <el-option
          :label="t('permission.option.moduleConfig')"
          value="config"
        />
      </el-select>
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
        type="danger"
        :icon="Delete"
        :disabled="selectedRows.length === 0"
        @click="handleBatchDelete"
      >
        {{ t('common.action.batchDelete') }}
      </el-button>
      <el-button
        :icon="Refresh"
        @click="fetchList"
      >
        {{ t('common.action.refresh') }}
      </el-button>
    </template>

    <template #col-module="{ row }">
      <StatusTag
        :status="row.module"
        :status-map="MODULE_STATUS_MAP"
      />
    </template>

    <template #col-status="{ row }">
      <StatusTag :status="row.status" />
    </template>

    <template #col-createTime="{ row }">
      {{ formatDate(row.createTime) }}
    </template>

    <template #col-updateTime="{ row }">
      {{ formatDate(row.updateTime) }}
    </template>

    <template #col-actions="{ row }">
      <el-button
        link
        type="primary"
        size="small"
        @click="openDrawer('view', row)"
      >
        {{ t('common.action.view') }}
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

  <PermissionFormDrawer
    v-model="drawerVisible"
    :mode="drawerMode"
    :data="editingRow"
    @success="onFormSuccess"
  />
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Delete, Refresh } from '@element-plus/icons-vue'
import { SearchTable, StatusTag } from '@/app/components'
import { useCrud } from '@/app/composables/useCrud'
import { formatDate } from '@/lib/format'
import { t } from '@/lib/i18n'
import type { ColumnDef } from '@/app/components/SearchTable/types'
import {
  fetchPermissionList,
  deletePermission,
  batchDeletePermissions,
  type PermissionInfo,
  type PermissionSearchRequest,
} from '../../permission/api'
import PermissionFormDrawer from './PermissionFormDrawer.vue'
import { MODULE_STATUS_MAP } from '@/app/constants/enums'

const {
  listData,
  loading,
  pagination,
  searchForm,
  selectedRows,
  fetchList,
  handleSearch,
  handleReset,
  handlePageChange,
  handleSelectionChange,
  handleDelete,
  handleBatchDelete,
} = useCrud<PermissionInfo>({
  fetch: (params) => fetchPermissionList(params as unknown as PermissionSearchRequest),
  remove: deletePermission,
  batchRemove: batchDeletePermissions,
  defaultSearchForm: { keyword: '', module: '', status: '' },
  pageSize: 10,
})

// SearchTable emit 的 selectionChange 类型是 Record<string, unknown>[]，需断言回 PermissionInfo[]
const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as PermissionInfo[])
}

// SearchTable 的 data/selectedRows prop 类型是 Record<string, unknown>[]，
// PermissionInfo 接口无索引签名，需 unknown 中转断言
const tableData = computed(() => listData.value as unknown as Record<string, unknown>[])
const tableSelectedRows = computed(() => selectedRows.value as unknown as Record<string, unknown>[])

const columns = computed<ColumnDef[]>(() => [
  { prop: 'name', label: t('permission.field.name'), minWidth: 140 },
  { prop: 'code', label: t('permission.field.code'), minWidth: 140 },
  { prop: 'module', label: t('permission.field.module'), minWidth: 110, slot: 'module' },
  { prop: 'description', label: t('common.column.description'), minWidth: 200 },
  { prop: 'status', label: t('common.column.status'), minWidth: 90, slot: 'status' },
  { prop: 'createTime', label: t('common.column.createTime'), minWidth: 170, slot: 'createTime' },
  { prop: 'updateTime', label: t('common.column.updateTime'), minWidth: 170, slot: 'updateTime' },
  { prop: 'actions', label: t('common.column.actions'), width: 200, fixed: 'right', slot: 'actions' },
])

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit' | 'view'>('add')
const editingRow = ref<PermissionInfo | null>(null)

const openDrawer = (mode: 'add' | 'edit' | 'view', row?: PermissionInfo) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}

const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}

onMounted(fetchList)
</script>
