<template>
  <FormDrawer
    v-model="visible"
    :title="drawerTitle"
    :mode="mode"
    :form-data="formData"
    :fields="fields"
    :rules="rules"
    :loading="submitting"
    width="600px"
    @submit="handleSubmit"
  />
</template>

<script lang="ts" setup>
import { ref, watch, reactive, computed } from 'vue'
import { FormDrawer } from '@/app/components'
import type { FormField, FormDrawerMode, TreeNodeData } from '@/app/components/FormDrawer/types'
import { ElMessage } from 'element-plus'
import {
  createMenu,
  updateMenu,
  type MenuInfo,
  type MenuCreateRequest,
} from '../api'
import { getCommonStatusOptions } from '@/app/constants/enums'
import { t } from '@/lib/i18n'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: MenuInfo | null
  /** 父菜单预设（点击"新增子菜单"时传入） */
  parent: MenuInfo | null
  /** 完整菜单树，用于 treeSelect 选择父节点 */
  treeData: MenuInfo[]
}>()

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  success: []
}>()

const visible = ref(props.modelValue)
const submitting = ref(false)

watch(() => props.modelValue, (v) => {
  visible.value = v
  if (v) initForm()
})
watch(visible, (v) => emit('update:modelValue', v))

const formData = reactive<MenuCreateRequest>({
  parentId: null,
  name: '',
  path: '',
  component: '',
  icon: '',
  sort: 0,
  status: 'active',
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return props.parent ? t('menu.addSubTitle', { name: props.parent.name }) : t('menu.addTitle')
  if (props.mode === 'edit') return t('menu.editTitle')
  return t('menu.viewTitle')
})

// treeSelect 数据需要 {id, label, children} 结构
const treeSelectData = computed<TreeNodeData[]>(() => {
  const transform = (nodes: MenuInfo[]): TreeNodeData[] =>
    nodes.map((n) => ({
      id: n.id,
      label: n.name,
      children: n.children && n.children.length > 0 ? transform(n.children) : undefined,
    }))
  return transform(props.treeData)
})

const fields = computed<FormField[]>(() => [
  {
    prop: 'parentId',
    label: t('menu.field.parentId'),
    type: 'treeSelect',
    span: 24,
    treeData: treeSelectData.value,
    treeProps: { label: 'label', children: 'children' },
    placeholder: t('menu.placeholder.noParent'),
    dependencies: [
      { trigger: 'mode', show: (_, ctx) => ctx.mode === 'add' && !props.parent }
    ],
  },
  { prop: 'name', label: t('menu.field.name'), type: 'input', span: 12 },
  { prop: 'path', label: t('menu.field.path'), type: 'input', span: 12 },
  { prop: 'component', label: t('menu.field.component'), type: 'input', span: 12, placeholder: t('menu.placeholder.component') },
  { prop: 'icon', label: t('menu.field.icon'), type: 'input', span: 12, placeholder: t('menu.placeholder.icon') },
  { prop: 'sort', label: t('menu.field.sort'), type: 'number', span: 12, default: 0 },
  {
    prop: 'status',
    label: t('menu.field.status'),
    type: 'radio',
    span: 12,
    options: getCommonStatusOptions(),
  },
])

const rules = computed(() => ({
  name: [{ required: true, message: t('menu.validation.nameRequired'), trigger: 'blur' }],
  path: [{ required: true, message: t('menu.validation.pathRequired'), trigger: 'blur' }],
  status: [{ required: true, message: t('menu.validation.statusRequired'), trigger: 'change' }],
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      parentId: props.data.parentId,
      name: props.data.name,
      path: props.data.path,
      component: props.data.component || '',
      icon: props.data.icon || '',
      sort: props.data.sort,
      status: props.data.status,
    })
  } else if (props.mode === 'add' && props.parent) {
    Object.assign(formData, {
      parentId: props.parent.id,
      name: '',
      path: '',
      component: '',
      icon: '',
      sort: 0,
      status: 'active',
    })
  } else {
    Object.assign(formData, {
      parentId: null,
      name: '',
      path: '',
      component: '',
      icon: '',
      sort: 0,
      status: 'active',
    })
  }
}

const handleSubmit = async (data: Record<string, unknown>) => {
  submitting.value = true
  try {
    const payload = { ...formData, ...data } as MenuCreateRequest
    if (props.mode === 'add') {
      await createMenu(payload)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updateMenu(props.data!.id, payload)
      ElMessage.success(t('common.message.updateSuccess'))
    }
    emit('success')
    emit('update:modelValue', false)
  } catch {
    // 失败由 http 拦截器提示，吞掉避免冒泡成 unhandled rejection
  } finally {
    submitting.value = false
  }
}
</script>
