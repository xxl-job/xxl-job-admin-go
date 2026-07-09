<!-- 这是使用通用组件实现 CRUD 的最小样例。
  展示：SearchTable + useCrud + PageContainer + FormDrawer 四件套的标准用法。
  业务模块可直接参考此模式复用。
-->
<template>
  <SearchTable
    :title="t('crud.title')"
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
        :placeholder="t('crud.field.name')"
        clearable
        style="width: 220px"
        @keyup.enter="handleSearch"
      />
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

  <CrudFormDrawer
    v-model="drawerVisible"
    :mode="drawerMode"
    :data="editingRow"
    @success="onFormSuccess"
  />
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Delete, Refresh } from '@element-plus/icons-vue'
import { SearchTable } from '@/app/components'
import { useCrud } from '@/app/composables/useCrud'
import type { ColumnDef } from '@/app/components/SearchTable/types'
import {
  fetchCrud,
  deleteCrudItem,
  batchDeleteCrudItems,
  type item as CrudItem
} from '../api'
import CrudFormDrawer from './CrudFormDrawer.vue'
import { t } from '@/lib/i18n'

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
  handleBatchDelete
} = useCrud<CrudItem>({
  fetch: (params) =>
    fetchCrud({
      name: (params.keyword as string) || '',
      current: Number(params.page ?? 1),
      size: Number(params.size ?? 10),
    }),
  remove: deleteCrudItem,
  batchRemove: batchDeleteCrudItems,
  defaultSearchForm: { keyword: '' },
  pageSize: 10
})

const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as CrudItem[])
}

const tableData = computed(
  () => listData.value as unknown as Record<string, unknown>[]
)
const tableSelectedRows = computed(
  () => selectedRows.value as unknown as Record<string, unknown>[]
)

const columns = computed<ColumnDef[]>(() => [
  { prop: 'name', label: t('crud.field.name'), minWidth: 120 },
  { prop: 'province', label: t('crud.field.province'), minWidth: 100 },
  { prop: 'city', label: t('crud.field.city'), minWidth: 100 },
  { prop: 'address', label: t('crud.field.address'), minWidth: 240 },
  { prop: 'zip', label: t('crud.field.zip'), width: 100 },
  { prop: 'date', label: t('crud.field.date'), width: 120 },
  {
    prop: 'actions',
    label: t('common.column.actions'),
    width: 180,
    fixed: 'right',
    slot: 'actions'
  }
])

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit' | 'view'>('add')
const editingRow = ref<CrudItem | null>(null)

const openDrawer = (mode: 'add' | 'edit' | 'view', row?: CrudItem) => {
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
