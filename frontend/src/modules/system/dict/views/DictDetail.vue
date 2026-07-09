<template>
  <el-card
    shadow="never"
    class="detail-card"
  >
    <template #header>
      <div class="detail-header">
        <div
          v-if="node"
          class="header-left"
        >
          <el-breadcrumb
            separator="/"
            class="breadcrumb"
          >
            <el-breadcrumb-item
              v-for="p in nodePath"
              :key="p.id"
            >
              {{ p.name }}
            </el-breadcrumb-item>
          </el-breadcrumb>
          <el-tag
            size="small"
            :type="levelTagMap[node.level]"
          >
            {{ levelLabelMap[node.level] }}
          </el-tag>
        </div>
        <span v-else>{{ title }}</span>
        <div
          v-if="node"
          class="header-actions"
        >
          <el-button
            v-if="node.level < 3"
            size="small"
            :icon="Plus"
            @click="handleAddChild"
          >
            {{ t('common.action.create') }}{{ levelLabelMap[node.level + 1] }}
          </el-button>
          <el-button
            size="small"
            :icon="Edit"
            @click="handleEdit"
          >
            {{ t('common.action.edit') }}
          </el-button>
          <el-button
            size="small"
            type="danger"
            :icon="Delete"
            @click="handleDelete"
          >
            {{ t('common.action.delete') }}
          </el-button>
        </div>
      </div>
    </template>

    <!-- 空状态引导 -->
    <div
      v-if="!node"
      class="empty-state"
    >
      <el-empty
        :description="t('dict.emptyDesc')"
        :image-size="100"
      >
        <el-button
          type="primary"
          :icon="Plus"
          @click="$emit('addRoot')"
        >
          {{ t('dict.addFirstCategory') }}
        </el-button>
      </el-empty>
    </div>

    <!-- 分类详情 -->
    <template v-else-if="node.level === 1">
      <el-descriptions
        :column="2"
        border
        size="small"
      >
        <el-descriptions-item :label="t('dict.field.categoryName')">
          {{ node.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.categoryCode')">
          {{ node.code }}
        </el-descriptions-item>
        <el-descriptions-item
          :label="t('common.column.description')"
          :span="2"
        >
          {{ node.description || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.status')">
          <el-tag
            :type="node.status === 'active' ? 'success' : 'danger'"
            size="small"
          >
            {{ node.status === 'active' ? t('common.status.enable') : t('common.status.disable') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.createTime')">
          {{ formatDate(node.createTime) }}
        </el-descriptions-item>
      </el-descriptions>

      <DictChildrenPreview
        v-if="node.children?.length || true"
        :items="node.children || []"
        :title="t('dict.children.containDict')"
        :parent-label="t('dict.children.category')"
        :child-label="t('dict.children.dict')"
        @add-child="handleAddChild"
      />
    </template>

    <!-- 字典详情 -->
    <template v-else-if="node.level === 2">
      <el-descriptions
        :column="2"
        border
        size="small"
      >
        <el-descriptions-item :label="t('dict.field.dictName')">
          {{ node.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.dictCode')">
          {{ node.code }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.category')">
          {{ categoryName }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.column.description')">
          {{ node.description || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.status')">
          <el-tag
            :type="node.status === 'active' ? 'success' : 'danger'"
            size="small"
          >
            {{ node.status === 'active' ? t('common.status.enable') : t('common.status.disable') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.createTime')">
          {{ formatDate(node.createTime) }}
        </el-descriptions-item>
      </el-descriptions>

      <DictChildrenPreview
        :items="node.children || []"
        :title="t('dict.children.containItem')"
        :parent-label="t('dict.children.dict')"
        :child-label="t('dict.children.item')"
        show-value-column
        @add-child="handleAddChild"
      />
    </template>

    <!-- 字典项详情 -->
    <template v-else>
      <el-descriptions
        :column="2"
        border
        size="small"
      >
        <el-descriptions-item :label="t('dict.field.itemName')">
          {{ node.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.itemCode')">
          {{ node.code }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.parentDict')">
          {{ dictName }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.dictValue')">
          {{ node.value }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.sort')">
          {{ node.sort }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.status')">
          <el-tag
            :type="node.status === 'active' ? 'success' : 'danger'"
            size="small"
          >
            {{ node.status === 'active' ? t('common.status.enable') : t('common.status.disable') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item
          :label="t('common.column.description')"
          :span="2"
        >
          {{ node.description || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.createTime')">
          {{ formatDate(node.createTime) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dict.field.updateTime')">
          {{ formatDate(node.updateTime) }}
        </el-descriptions-item>
      </el-descriptions>
    </template>
  </el-card>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import type { DictTreeNode } from './hooks/useDictTree'
import { formatDate } from '@/lib/format'
import { t } from '@/lib/i18n'
import DictChildrenPreview from './components/DictChildrenPreview.vue'

const props = defineProps<{
  node: DictTreeNode | null
  treeData: DictTreeNode[]
  categoryName?: string
  dictName?: string
}>()

const emit = defineEmits<{
  (e: 'edit', node: DictTreeNode): void
  (e: 'delete', node: DictTreeNode): void
  (e: 'addChild', node: DictTreeNode): void
  (e: 'addRoot'): void
}>()

const levelLabelMap = computed<Record<number, string>>(() => ({
  1: t('dict.level.1'),
  2: t('dict.level.2'),
  3: t('dict.level.3'),
}))

const levelTagMap: Record<number, string> = {
  1: 'primary',
  2: 'success',
  3: 'warning',
}

const title = computed(() => {
  if (!props.node) return t('dict.selectNode')
  if (props.node.level === 1) return t('dict.categoryDetail')
  if (props.node.level === 2) return t('dict.dictDetail')
  return t('dict.itemDetail')
})

// 查找父节点路径
function findParentPath(
  nodes: DictTreeNode[],
  targetId: string,
  path: DictTreeNode[] = []
): DictTreeNode[] {
  for (const n of nodes) {
    if (n.id === targetId) return [...path, n]
    if (n.children?.length) {
      const result = findParentPath(n.children, targetId, [...path, n])
      if (result.length) return result
    }
  }
  return []
}

const nodePath = computed(() => {
  if (!props.node) return []
  return findParentPath(props.treeData, props.node.id)
})

const handleEdit = () => {
  if (props.node) emit('edit', props.node)
}

const handleDelete = () => {
  if (props.node) emit('delete', props.node)
}

const handleAddChild = () => {
  if (props.node) emit('addChild', props.node)
}
</script>

<style scoped>
.detail-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.detail-card :deep(.el-card__body) {
  flex: 1;
  overflow-y: auto;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.breadcrumb {
  flex: 1;
  min-width: 0;
}

.breadcrumb :deep(.el-breadcrumb__inner) {
  font-weight: 600;
}

.empty-state {
  padding: 40px 20px;
}
</style>
