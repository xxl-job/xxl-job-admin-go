<template>
  <el-drawer
    :model-value="modelValue"
    :title="t('role.permissionConfigTitle')"
    size="50%"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div
      v-loading="loading"
      class="permission-config"
    >
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="{ label: 'name', children: 'children' }"
        :default-checked-keys="checkedKeys"
        show-checkbox
        node-key="id"
        check-strictly
        class="permission-tree"
      />
    </div>
    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">
        {{ t('common.action.cancel') }}
      </el-button>
      <el-button
        type="primary"
        :loading="saving"
        @click="handleSave"
      >
        {{ t('common.action.save') }}
      </el-button>
    </template>
  </el-drawer>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted } from 'vue'
import type { ElTree } from 'element-plus'
import { ElMessage } from 'element-plus'
import { t } from '@/lib/i18n'
import {
  fetchRolePermissions,
  setRolePermissions,
} from '../../role/api'
import {
  fetchAllPermissions,
  type PermissionInfo,
} from '../../permission/api'

interface PermissionTreeNode {
  id: string
  name: string
  children?: PermissionTreeNode[]
}

const MODULE_LABELS: Record<string, string> = {
  system: 'permission.option.moduleSystem',
  user: 'permission.option.moduleUser',
  role: 'permission.option.moduleRole',
  permission: 'permission.option.modulePermission',
  dict: 'permission.option.moduleDict',
  config: 'permission.option.moduleConfig',
}

const props = defineProps<{
  modelValue: boolean
  roleId: string | null
}>()

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
}>()

const loading = ref(false)
const saving = ref(false)
const treeRef = ref<InstanceType<typeof ElTree>>()
const treeData = ref<PermissionTreeNode[]>([])
const checkedKeys = ref<string[]>([])

const buildTree = (permissions: PermissionInfo[]): PermissionTreeNode[] => {
  const moduleMap = new Map<string, PermissionInfo[]>()
  permissions.forEach((p) => {
    if (!moduleMap.has(p.module)) moduleMap.set(p.module, [])
    moduleMap.get(p.module)!.push(p)
  })
  const tree: PermissionTreeNode[] = []
  for (const [module, items] of moduleMap.entries()) {
    tree.push({
      id: module,
      name: t(MODULE_LABELS[module]) || module,
      children: items.map((p) => ({ id: p.id, name: p.name })),
    })
  }
  return tree
}

const loadData = async () => {
  if (!props.roleId) return
  loading.value = true
  try {
    const [allPermissions, rolePermissions] = await Promise.all([
      fetchAllPermissions(),
      fetchRolePermissions(props.roleId),
    ])
    treeData.value = buildTree(allPermissions)
    checkedKeys.value = rolePermissions
  } catch {
    // 失败由 http 拦截器提示
  } finally {
    loading.value = false
  }
}

watch(() => props.modelValue, (v) => {
  if (v) loadData()
})

onMounted(() => {
  if (props.modelValue) loadData()
})

const handleSave = async () => {
  if (!props.roleId || !treeRef.value) return
  saving.value = true
  try {
    const keys = treeRef.value.getCheckedKeys() as string[]
    // 过滤掉模块节点（id 是 module 名）
    const permissionIds = keys.filter((k) => !Object.prototype.hasOwnProperty.call(MODULE_LABELS, k))
    await setRolePermissions(props.roleId, permissionIds)
    ElMessage.success(t('role.permissionSaveSuccess'))
    emit('update:modelValue', false)
  } catch {
    // 失败由 http 拦截器提示
  } finally {
    saving.value = false
  }
}
</script>

<style lang="scss" scoped>
.permission-config {
  min-height: 100%;
}
</style>
