<template>
  <SearchTable
    :title="t('role.title')"
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
        :placeholder="t('role.searchKeyword')"
        clearable
        style="width: 220px"
        @keyup.enter="handleSearch"
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
        type="success"
        size="small"
        @click="openPermissionDrawer(row)"
      >
        {{ t('role.permissionLabel') }}
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

  <RoleFormDrawer
    v-model="drawerVisible"
    :mode="drawerMode"
    :data="editingRow"
    @success="onFormSuccess"
  />

  <RolePermissionDrawer
    v-model="permissionDrawerVisible"
    :role-id="currentRoleId"
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
  fetchRoleList,
  deleteRole,
  batchDeleteRoles,
  type RoleInfo,
  type RoleSearchRequest,
} from '../../role/api'
import RoleFormDrawer from './RoleFormDrawer.vue'
import RolePermissionDrawer from './RolePermissionDrawer.vue'

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
} = useCrud<RoleInfo>({
  fetch: (params) => fetchRoleList(params as unknown as RoleSearchRequest),
  remove: deleteRole,
  batchRemove: batchDeleteRoles,
  defaultSearchForm: { keyword: '', status: '' },
  pageSize: 10,
})

// SearchTable emit 的 selectionChange 类型是 Record<string, unknown>[]，需断言回 RoleInfo[]
const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as RoleInfo[])
}

// SearchTable 的 data/selectedRows prop 类型是 Record<string, unknown>[]，
// RoleInfo 接口无索引签名，需 unknown 中转断言
const tableData = computed(() => listData.value as unknown as Record<string, unknown>[])
const tableSelectedRows = computed(() => selectedRows.value as unknown as Record<string, unknown>[])

const columns = computed<ColumnDef[]>(() => [
  { prop: 'name', label: t('role.field.name'), minWidth: 140 },
  { prop: 'code', label: t('role.field.code'), minWidth: 140 },
  { prop: 'description', label: t('common.column.description'), minWidth: 200 },
  { prop: 'status', label: t('common.column.status'), minWidth: 90, slot: 'status' },
  { prop: 'createTime', label: t('common.column.createTime'), minWidth: 170, slot: 'createTime' },
  { prop: 'updateTime', label: t('common.column.updateTime'), minWidth: 170, slot: 'updateTime' },
  { prop: 'actions', label: t('common.column.actions'), width: 240, fixed: 'right', slot: 'actions' },
])

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit' | 'view'>('add')
const editingRow = ref<RoleInfo | null>(null)
const permissionDrawerVisible = ref(false)
const currentRoleId = ref<string | null>(null)

const openDrawer = (mode: 'add' | 'edit' | 'view', row?: RoleInfo) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}

const openPermissionDrawer = (row: RoleInfo) => {
  currentRoleId.value = row.id
  permissionDrawerVisible.value = true
}

const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}

onMounted(fetchList)
</script>
