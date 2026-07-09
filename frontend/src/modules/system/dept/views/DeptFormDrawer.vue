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
  >
    <!-- 上级部门：复用 DeptSelector，免去本组件手动加载树 -->
    <template #field-parentId>
      <DeptSelector
        v-model="formData.parentId"
        :disabled="mode === 'view'"
        :placeholder="t('dept.placeholder.noParent')"
      />
    </template>
  </FormDrawer>
</template>

<script lang="ts" setup>
import { ref, watch, reactive, computed } from 'vue'
import { FormDrawer, DeptSelector } from '@/app/components'
import type { FormField, FormDrawerMode } from '@/app/components/FormDrawer/types'
import { ElMessage } from 'element-plus'
import {
  createDept,
  updateDept,
  type DeptInfo,
  type DeptCreateRequest,
} from '../api'
import { getCommonStatusOptions } from '@/app/constants/enums'
import { t } from '@/lib/i18n'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: DeptInfo | null
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

const formData = reactive<DeptCreateRequest>({
  name: '',
  parentId: '',
  leader: '',
  phone: '',
  email: '',
  sort: 0,
  status: 'active',
})

const drawerTitle = computed(() => {
  if (props.mode === 'edit') return t('dept.editTitle')
  return t('dept.addTitle')
})

// parentId 由 #field-parentId 插槽内的 DeptSelector 渲染；
// 保留 treeSelect 类型占位以保证 label/span 正常计算。
const fields = computed<FormField[]>(() => [
  { prop: 'parentId', label: t('dept.field.parentId'), type: 'treeSelect' },
  { prop: 'name', label: t('dept.field.name'), type: 'input' },
  { prop: 'leader', label: t('dept.field.leader'), type: 'input' },
  { prop: 'phone', label: t('dept.field.phone'), type: 'input' },
  { prop: 'email', label: t('dept.field.email'), type: 'input' },
  { prop: 'sort', label: t('dept.field.sort'), type: 'number' },
  {
    prop: 'status',
    label: t('dept.field.status'),
    type: 'radio',
    options: getCommonStatusOptions(),
  },
])

const rules = computed(() => ({
  name: [{ required: true, message: t('dept.validation.nameRequired'), trigger: 'blur' }],
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      name: props.data.name,
      parentId: props.data.parentId,
      leader: props.data.leader,
      phone: props.data.phone,
      email: props.data.email,
      sort: props.data.sort,
      status: props.data.status,
    })
  } else {
    // 新增：若从某行触发，预填该行为上级
    Object.assign(formData, {
      name: '',
      parentId: props.data?.id || '',
      leader: '',
      phone: '',
      email: '',
      sort: 0,
      status: 'active',
    })
  }
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (props.mode === 'edit') {
      await updateDept(props.data!.id, formData)
      ElMessage.success(t('common.message.updateSuccess'))
    } else {
      await createDept(formData)
      ElMessage.success(t('common.message.createSuccess'))
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
