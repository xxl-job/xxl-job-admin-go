<template>
  <FormDrawer
    v-model="visible"
    :title="drawerTitle"
    :mode="mode"
    :form-data="formData"
    :fields="fields"
    :rules="rules"
    :loading="submitting"
    width="500px"
    @submit="handleSubmit"
  />
</template>

<script lang="ts" setup>
import { ref, watch, reactive, computed } from 'vue'
import { FormDrawer } from '@/app/components'
import type { FormField, FormDrawerMode } from '@/app/components/FormDrawer/types'
import { ElMessage } from 'element-plus'
import {
  createPermission,
  updatePermission,
  type PermissionInfo,
  type PermissionCreateRequest,
} from '../../permission/api'
import { getCommonStatusOptions } from '@/app/constants/enums'
import { t } from '@/lib/i18n'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: PermissionInfo | null
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

const formData = reactive<PermissionCreateRequest>({
  name: '',
  code: '',
  description: '',
  module: '',
  status: 'active',
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return t('permission.addTitle')
  if (props.mode === 'edit') return t('permission.editTitle')
  return t('permission.viewTitle')
})

const fields = computed<FormField[]>(() => [
  { prop: 'name', label: t('permission.field.name'), type: 'input', span: 12 },
  { prop: 'code', label: t('permission.field.code'), type: 'input', span: 12 },
  {
    prop: 'module',
    label: t('permission.field.module'),
    type: 'select',
    span: 12,
    options: [
      { label: t('permission.option.moduleSystem'), value: 'system' },
      { label: t('permission.option.moduleUser'), value: 'user' },
      { label: t('permission.option.moduleRole'), value: 'role' },
      { label: t('permission.option.modulePermission'), value: 'permission' },
      { label: t('permission.option.moduleDict'), value: 'dict' },
      { label: t('permission.option.moduleConfig'), value: 'config' },
    ],
  },
  {
    prop: 'status',
    label: t('permission.field.status'),
    type: 'radio',
    span: 12,
    options: getCommonStatusOptions(),
  },
  { prop: 'description', label: t('permission.field.description'), type: 'textarea', span: 24 },
])

const rules = computed(() => ({
  name: [{ required: true, message: t('permission.validation.nameRequired'), trigger: 'blur' }],
  code: [{ required: true, message: t('permission.validation.codeRequired'), trigger: 'blur' }],
  module: [{ required: true, message: t('permission.validation.moduleRequired'), trigger: 'change' }],
  status: [{ required: true, message: t('permission.validation.statusRequired'), trigger: 'change' }],
  description: [{ max: 200, message: t('permission.validation.descriptionMax'), trigger: 'blur' }],
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      name: props.data.name,
      code: props.data.code,
      description: props.data.description,
      module: props.data.module,
      status: props.data.status,
    })
  } else {
    Object.assign(formData, {
      name: '',
      code: '',
      description: '',
      module: '',
      status: 'active',
    })
  }
}

const handleSubmit = async (data: Record<string, unknown>) => {
  submitting.value = true
  try {
    const payload = { ...formData, ...data } as PermissionCreateRequest
    if (props.mode === 'add') {
      await createPermission(payload)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updatePermission(props.data!.id, payload)
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
