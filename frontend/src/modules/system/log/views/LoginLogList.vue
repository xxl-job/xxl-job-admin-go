<template>
  <SearchTable
    :title="t('log.loginTitle')"
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
        :placeholder="t('log.loginKeyword')"
        clearable
        style="width: 200px"
        @keyup.enter="handleSearch"
      />
      <el-select
        v-model="searchForm.status"
        clearable
        :placeholder="t('common.column.status')"
        style="width: 120px"
      >
        <el-option
          :label="t('common.status.success')"
          value="success"
        />
        <el-option
          :label="t('common.status.failed')"
          value="failed"
        />
      </el-select>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        :range-separator="t('log.dateSeparator')"
        :start-placeholder="t('log.dateStart')"
        :end-placeholder="t('log.dateEnd')"
        style="width: 260px"
        @change="handleDateChange"
      />
    </template>

    <template #actions>
      <el-button
        type="danger"
        :icon="Delete"
        :disabled="selectedRows.length === 0"
        @click="handleBatchDelete"
      >
        {{ t('common.action.batchDelete') }}
      </el-button>
      <el-button
        type="warning"
        :icon="Delete"
        @click="handleClear"
      >
        {{ t('log.clearLog') }}
      </el-button>
      <el-button
        :icon="Download"
        @click="handleExport"
      >
        {{ t('common.action.export') }}
      </el-button>
      <el-button
        :icon="Refresh"
        @click="fetchList"
      >
        {{ t('common.action.refresh') }}
      </el-button>
    </template>

    <template #col-username="{ row }">
      <el-tag
        :type="row.status === 'success' ? 'success' : 'danger'"
        size="small"
      >
        {{ row.username }}
      </el-tag>
    </template>

    <template #col-status="{ row }">
      <el-tag
        :type="row.status === 'success' ? 'success' : 'danger'"
        size="small"
      >
        {{ row.status === 'success' ? t('common.status.success') : t('common.status.failed') }}
      </el-tag>
    </template>

    <template #col-loginTime="{ row }">
      {{ formatDate(row.loginTime) }}
    </template>
  </SearchTable>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue'
import { Delete, Refresh, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { confirmService } from '@/lib/confirm'
import { SearchTable } from '@/app/components'
import { useCrud } from '@/app/composables/useCrud'
import { formatDate } from '@/lib/format'
import { downloadCsv } from '@/lib/file'
import { t } from '@/lib/i18n'
import type { ColumnDef } from '@/app/components/SearchTable/types'
import {
  fetchLoginLogList,
  deleteLoginLog,
  batchDeleteLoginLogs,
  clearLoginLogs,
  exportLoginLogs,
  type LoginLogInfo,
  type LogSearchRequest,
} from '../api'

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
  handleBatchDelete,
} = useCrud<LoginLogInfo>({
  fetch: (params) => fetchLoginLogList(params as unknown as LogSearchRequest),
  remove: deleteLoginLog,
  batchRemove: batchDeleteLoginLogs,
  defaultSearchForm: { keyword: '', status: '', startTime: '', endTime: '' },
  pageSize: 20,
})

const dateRange = ref<[string, string] | null>(null)

const handleDateChange = (dates: [string, string] | null) => {
  // searchForm 是 reactive（非 ref），无 .value；直接赋值保持响应性
  if (dates) {
    searchForm.startTime = dates[0]
    searchForm.endTime = dates[1]
  } else {
    searchForm.startTime = ''
    searchForm.endTime = ''
  }
}

const handleClear = async () => {
  const confirmed = await confirmService.showConfirm(t('log.confirmClearLogin'))
  if (!confirmed) return
  try {
    await clearLoginLogs()
    ElMessage.success(t('log.clearSuccess'))
    fetchList()
  } catch {
    // 失败由 http 拦截器提示
  }
}

const handleExport = async () => {
  try {
    const csv = await exportLoginLogs()
    downloadCsv(csv, t('log.loginFileName'))
    ElMessage.success(t('common.message.exportSuccess'))
  } catch {
    // 失败由 http 拦截器提示
  }
}

const onSelectionChange = (rows: Record<string, unknown>[]) => {
  handleSelectionChange(rows as unknown as LoginLogInfo[])
}

const tableData = computed(() => listData.value as unknown as Record<string, unknown>[])
const tableSelectedRows = computed(() => selectedRows.value as unknown as Record<string, unknown>[])

const columns = computed<ColumnDef[]>(() => [
  { prop: 'username', label: t('user.field.username'), minWidth: 120, slot: 'username' },
  { prop: 'ip', label: t('log.field.ip'), minWidth: 140 },
  { prop: 'location', label: t('log.field.loginLocation'), minWidth: 150 },
  { prop: 'browser', label: t('log.field.browser'), minWidth: 140 },
  { prop: 'os', label: t('log.field.os'), minWidth: 120 },
  { prop: 'status', label: t('common.column.status'), minWidth: 90, slot: 'status' },
  { prop: 'message', label: t('log.field.message'), minWidth: 180 },
  { prop: 'loginTime', label: t('log.field.loginTime'), minWidth: 180, slot: 'loginTime' },
])

onMounted(fetchList)
</script>
