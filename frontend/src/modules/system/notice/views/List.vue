<template>
  <SearchTable
    :title="t('notice.title')"
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
        :placeholder="t('notice.searchKeyword')"
        clearable
        style="width: 220px"
        @keyup.enter="handleSearch"
      />
      <el-select
        v-model="searchForm.type"
        clearable
        :placeholder="t('notice.field.type')"
        style="width: 120px"
      >
        <el-option
          :label="t('notice.option.typeAnnouncement')"
          value="notice"
        />
        <el-option
          :label="t('notice.option.typeNotice')"
          value="notification"
        />
        <el-option
          :label="t('notice.option.typeTodo')"
          value="todo"
        />
      </el-select>
      <el-select
        v-model="searchForm.status"
        clearable
        :placeholder="t('common.column.status')"
        style="width: 120px"
      >
        <el-option
          :label="t('notice.option.statusPublished')"
          value="published"
        />
        <el-option
          :label="t('notice.option.statusDraft')"
          value="draft"
        />
        <el-option
          :label="t('notice.option.statusRevoked')"
          value="revoked"
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

    <template #col-type="{ row }">
      <el-tag
        :type="row.type === 'notice' ? 'danger' : row.type === 'notification' ? 'warning' : 'info'"
        size="small"
      >
        {{ row.type === 'notice' ? t('notice.option.typeAnnouncement') : row.type === 'notification' ? t('notice.option.typeNotice') : t('notice.option.typeTodo') }}
      </el-tag>
    </template>

    <template #col-status="{ row }">
      <el-tag
        :type="row.status === 'published' ? 'success' : row.status === 'draft' ? 'info' : 'danger'"
        size="small"
      >
        {{ row.status === 'published' ? t('notice.option.statusPublished') : row.status === 'draft' ? t('notice.option.statusDraft') : t('notice.option.statusRevoked') }}
      </el-tag>
    </template>

    <template #col-priority="{ row }">
      <el-tag
        :type="row.priority === 'high' ? 'danger' : row.priority === 'medium' ? 'warning' : 'info'"
        size="small"
      >
        {{ row.priority === 'high' ? t('notice.option.priorityHigh') : row.priority === 'medium' ? t('notice.option.priorityMedium') : t('notice.option.priorityLow') }}
      </el-tag>
    </template>

    <template #col-publishTime="{ row }">
      {{ formatDate(row.publishTime) }}
    </template>

    <template #col-actions="{ row }">
      <el-button
        v-if="row.status === 'draft'"
        link
        type="success"
        size="small"
        @click="handlePublish(row.id)"
      >
        {{ t('notice.actionPublish') }}
      </el-button>
      <el-button
        v-if="row.status === 'published'"
        link
        type="warning"
        size="small"
        @click="handleRevoke(row.id)"
      >
        {{ t('notice.actionRevoke') }}
      </el-button>
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

  <NoticeFormDrawer
    v-model="drawerVisible"
    :mode="drawerMode"
    :data="editingRow"
    @success="onFormSuccess"
  />
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Delete, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { confirmService } from '@/lib/confirm'
import { SearchTable } from '@/app/components'
import { useCrud } from '@/app/composables/useCrud'
import { formatDate } from '@/lib/format'
import { t } from '@/lib/i18n'
import type { ColumnDef } from '@/app/components/SearchTable/types'
import {
  fetchNoticeList,
  deleteNotice,
  batchDeleteNotices,
  publishNotice,
  revokeNotice,
  type NoticeInfo,
  type NoticeSearchRequest,
} from '../api'
import NoticeFormDrawer from './NoticeFormDrawer.vue'

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
} = useCrud<NoticeInfo>({
  fetch: (params) => fetchNoticeList(params as unknown as NoticeSearchRequest),
  remove: deleteNotice,
  batchRemove: batchDeleteNotices,
  defaultSearchForm: { keyword: '', type: '', status: '' },
  pageSize: 10,
})

const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as NoticeInfo[])
}

const tableData = computed(() => listData.value as unknown as Record<string, unknown>[])
const tableSelectedRows = computed(() => selectedRows.value as unknown as Record<string, unknown>[])

const columns = computed<ColumnDef[]>(() => [
  { prop: 'title', label: t('notice.field.title'), minWidth: 240 },
  { prop: 'type', label: t('notice.field.type'), minWidth: 100, slot: 'type' },
  { prop: 'priority', label: t('notice.field.priority'), minWidth: 100, slot: 'priority' },
  { prop: 'status', label: t('notice.field.status'), minWidth: 100, slot: 'status' },
  { prop: 'publisher', label: t('common.column.publisher'), minWidth: 100 },
  { prop: 'publishTime', label: t('common.column.publishTime'), minWidth: 170, slot: 'publishTime' },
  { prop: 'actions', label: t('common.column.actions'), width: 220, fixed: 'right', slot: 'actions' },
])

const drawerVisible = ref(false)
const drawerMode = ref<'add' | 'edit' | 'view'>('add')
const editingRow = ref<NoticeInfo | null>(null)

const openDrawer = (mode: 'add' | 'edit' | 'view', row?: NoticeInfo) => {
  drawerMode.value = mode
  editingRow.value = row ?? null
  drawerVisible.value = true
}

const onFormSuccess = () => {
  drawerVisible.value = false
  fetchList()
}

const handlePublish = async (id: string) => {
  const confirmed = await confirmService.showConfirm(t('notice.confirmPublish'))
  if (!confirmed) return
  try {
    await publishNotice(id)
    ElMessage.success(t('notice.publishSuccess'))
    fetchList()
  } catch {
    // 失败由 http 拦截器提示
  }
}

const handleRevoke = async (id: string) => {
  const confirmed = await confirmService.showConfirm(t('notice.confirmRevoke'))
  if (!confirmed) return
  try {
    await revokeNotice(id)
    ElMessage.success(t('notice.revokeSuccess'))
    fetchList()
  } catch {
    // 失败由 http 拦截器提示
  }
}

onMounted(fetchList)
</script>
