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
  createRole,
  updateRole,
  type RoleInfo,
  type RoleCreateRequest,
} from '../../role/api'
import { getCommonStatusOptions } from '@/app/constants/enums'
import { t } from '@/lib/i18n'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: RoleInfo | null
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

const formData = reactive<RoleCreateRequest>({
  name: '',
  code: '',
  description: '',
  status: 'active',
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return t('role.addTitle')
  if (props.mode === 'edit') return t('role.editTitle')
  return t('role.viewTitle')
})

const fields = computed<FormField[]>(() => [
  { prop: 'name', label: t('role.field.name'), type: 'input', span: 24 },
  { prop: 'code', label: t('role.field.code'), type: 'input', span: 24 },
  { prop: 'description', label: t('role.field.description'), type: 'textarea', span: 24 },
  {
    prop: 'status',
    label: t('role.field.status'),
    type: 'radio',
    span: 24,
    options: getCommonStatusOptions(),
  },
])

const rules = computed(() => ({
  name: [
    { required: true, message: t('role.validation.nameRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('role.validation.nameLength'), trigger: 'blur' },
  ],
  code: [
    { required: true, message: t('role.validation.codeRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('role.validation.codeLength'), trigger: 'blur' },
  ],
  description: [{ max: 200, message: t('role.validation.descriptionMax'), trigger: 'blur' }],
  status: [{ required: true, message: t('role.validation.statusRequired'), trigger: 'change' }],
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      name: props.data.name,
      code: props.data.code,
      description: props.data.description,
      status: props.data.status,
    })
  } else {
    Object.assign(formData, {
      name: '',
      code: '',
      description: '',
      status: 'active',
    })
  }
}

const handleSubmit = async (data: Record<string, unknown>) => {
  submitting.value = true
  try {
    const payload = { ...formData, ...data } as RoleCreateRequest
    if (props.mode === 'add') {
      await createRole(payload)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updateRole(props.data!.id, payload)
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
