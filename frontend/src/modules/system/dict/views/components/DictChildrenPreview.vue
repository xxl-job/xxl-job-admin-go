<template>
  <div class="dict-children-preview">
    <div class="preview-title">
      {{ title }}（{{ items.length }}）
    </div>
    <el-table
      v-if="items.length"
      :data="items.slice(0, 10)"
      size="small"
      border
      stripe
    >
      <el-table-column
        prop="name"
        :label="childLabel + t('dict.preview.nameSuffix')"
        min-width="120"
      />
      <el-table-column
        prop="code"
        :label="childLabel + t('dict.preview.codeSuffix')"
        min-width="110"
      />
      <el-table-column
        v-if="showValueColumn"
        prop="value"
        :label="t('dict.field.dictValue')"
        min-width="100"
      />
      <el-table-column
        prop="status"
        :label="t('common.column.status')"
        width="80"
        align="center"
      >
        <template #default="{ row }">
          <el-tag
            size="small"
            :type="row.status === 'active' ? 'success' : 'info'"
          >
            {{ row.status === 'active' ? t('common.status.enable') : t('common.status.disable') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="sort"
        :label="t('dict.field.sort')"
        width="80"
        align="right"
      />
    </el-table>
    <el-alert
      v-if="items.length > 10"
      type="info"
      size="small"
      show-icon
      class="preview-alert"
    >
      {{ t('dict.preview.more', { count: items.length - 10, child: childLabel }) }}
    </el-alert>
    <div
      v-else
      class="empty-children"
    >
      <el-empty
        :description="t('dict.preview.empty', { parent: parentLabel, child: childLabel })"
        :image-size="60"
      >
        <el-button
          size="small"
          type="primary"
          :icon="Plus"
          @click="$emit('addChild')"
        >
          {{ t('common.action.create') }}{{ childLabel }}
        </el-button>
      </el-empty>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { Plus } from '@element-plus/icons-vue'
import type { DictTreeNode } from '../hooks/useDictTree'
import { t } from '@/lib/i18n'

// 字典子节点预览：level 1（分类）和 level 2（字典）共用。
// 通过 childLabel / parentLabel 自适应文案；showValueColumn 控制是否展示"字典值"列。
withDefaults(
  defineProps<{
    items: DictTreeNode[]
    title: string
    parentLabel: string
    childLabel: string
    showValueColumn?: boolean
  }>(),
  { showValueColumn: false }
)

defineEmits<{ (e: 'addChild'): void }>()
</script>

<style scoped>
.dict-children-preview {
  margin-top: 20px;
}

.preview-title {
  margin-bottom: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.preview-alert {
  margin-top: 12px;
}

.empty-children {
  padding: 20px;
  border: 1px dashed var(--el-border-color);
  border-radius: 4px;
}
</style>
