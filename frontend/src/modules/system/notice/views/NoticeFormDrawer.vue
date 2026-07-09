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
import type { FormField, FormDrawerMode } from '@/app/components/FormDrawer/types'
import { ElMessage } from 'element-plus'
import { t } from '@/lib/i18n'
import {
  createNotice,
  updateNotice,
  type NoticeInfo,
  type NoticeCreateRequest,
} from '../api'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: NoticeInfo | null
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

const formData = reactive<NoticeCreateRequest>({
  title: '',
  content: '',
  type: 'notice',
  status: 'draft',
  priority: 'medium',
  publishTime: '',
  expireTime: '',
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return t('notice.addTitle')
  if (props.mode === 'edit') return t('notice.editTitle')
  return t('notice.viewTitle')
})

const fields = computed<FormField[]>(() => [
  { prop: 'title', label: t('notice.field.title'), type: 'input', span: 24 },
  {
    prop: 'type',
    label: t('notice.field.type'),
    type: 'select',
    span: 12,
    options: [
      { label: t('notice.option.typeAnnouncement'), value: 'announcement' },
      { label: t('notice.option.typeNotice'), value: 'notice' },
      { label: t('notice.option.typeTodo'), value: 'todo' },
    ],
  },
  {
    prop: 'status',
    label: t('notice.field.status'),
    type: 'select',
    span: 12,
    options: [
      { label: t('notice.option.statusDraft'), value: 'draft' },
      { label: t('notice.option.statusPublished'), value: 'published' },
      { label: t('notice.option.statusExpired'), value: 'expired' },
    ],
  },
  {
    prop: 'priority',
    label: t('notice.field.priority'),
    type: 'radio',
    span: 24,
    options: [
      { label: t('notice.option.priorityHigh'), value: 'high' },
      { label: t('notice.option.priorityMedium'), value: 'medium' },
      { label: t('notice.option.priorityLow'), value: 'low' },
    ],
  },
  {
    prop: 'content',
    label: t('notice.field.content'),
    type: 'textarea',
    span: 24,
  },
  { prop: 'publishTime', label: t('notice.field.publishTime'), type: 'date', span: 12 },
  { prop: 'expireTime', label: t('notice.field.expireTime'), type: 'date', span: 12 },
])

const rules = computed(() => ({
  title: [{ required: true, message: t('notice.validation.titleRequired'), trigger: 'blur' }],
  content: [{ required: true, message: t('notice.validation.contentRequired'), trigger: 'blur' }],
}))

const initForm = () => {
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    Object.assign(formData, {
      title: props.data.title,
      content: props.data.content,
      type: props.data.type,
      status: props.data.status,
      priority: props.data.priority,
      publishTime: props.data.publishTime || '',
      expireTime: props.data.expireTime || '',
    })
  } else {
    Object.assign(formData, {
      title: '',
      content: '',
      type: 'notice',
      status: 'draft',
      priority: 'medium',
      publishTime: '',
      expireTime: '',
    })
  }
}

const handleSubmit = async () => {
  submitting.value = true
  try {
    if (props.mode === 'add') {
      await createNotice(formData)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updateNotice(props.data!.id, formData)
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
