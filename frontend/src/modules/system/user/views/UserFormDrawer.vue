<template>
  <FormDrawer
    ref="formDrawerRef"
    v-model="visible"
    :title="drawerTitle"
    :mode="mode"
    :form-data="formData"
    :fields="fields"
    :rules="rules"
    :loading="submitting"
    width="600px"
    @submit="handleSubmit"
  >
    <!-- 角色多选：复用 RoleSelector（valueKey=code 对齐 UserProfile.roles），免去本组件手动加载角色列表 -->
    <template #field-roles>
      <RoleSelector
        v-model="formData.roles"
        multiple
        value-key="code"
        :disabled="mode === 'view'"
        :placeholder="t('user.placeholder.selectRole')"
      />
    </template>
  </FormDrawer>
</template>

<script lang="ts" setup>
import { ref, watch, reactive, computed } from 'vue'
import { FormDrawer, RoleSelector } from '@/app/components'
import type { FormField, FormDrawerMode } from '@/app/components/FormDrawer/types'
import { ElMessage } from 'element-plus'
import { t } from '@/lib/i18n'
import {
  createUser,
  updateUser,
  type UserInfo,
  type UserCreateRequest,
} from '../api'
import { getCommonStatusOptions } from '@/app/constants/enums'

const props = defineProps<{
  modelValue: boolean
  mode: FormDrawerMode
  data: UserInfo | null
}>()

const emit = defineEmits<{
  'update:modelValue': [v: boolean]
  success: []
}>()

const visible = ref(props.modelValue)
const submitting = ref(false)
const formDrawerRef = ref<InstanceType<typeof FormDrawer>>()

watch(() => props.modelValue, (v) => {
  visible.value = v
  if (v) initForm()
})
watch(visible, (v) => emit('update:modelValue', v))

const formData = reactive<UserCreateRequest & { confirmPassword: string }>({
  username: '',
  realName: '',
  email: '',
  phone: '',
  roles: [],
  status: 'active',
  password: '',
  confirmPassword: '',
})

const drawerTitle = computed(() => {
  if (props.mode === 'add') return t('user.addTitle')
  if (props.mode === 'edit') return t('user.editTitle')
  return t('user.viewTitle')
})

const fields = computed<FormField[]>(() => [
  { prop: 'username', label: t('user.field.username'), type: 'input', span: 12 },
  { prop: 'realName', label: t('user.field.realName'), type: 'input', span: 12 },
  { prop: 'email', label: t('user.field.email'), type: 'input', span: 12 },
  { prop: 'phone', label: t('user.field.phone'), type: 'input', span: 12 },
  {
    prop: 'roles',
    label: t('user.field.roles'),
    type: 'select',
    span: 12,
  },
  {
    prop: 'status',
    label: t('user.field.status'),
    type: 'select',
    span: 12,
    options: getCommonStatusOptions(),
  },
  {
    prop: 'password',
    label: t('user.field.password'),
    type: 'password',
    span: 12,
    placeholder: props.mode === 'edit' ? t('user.placeholder.passwordEdit') : t('user.placeholder.passwordAdd'),
    dependencies: [
      { trigger: 'mode', show: (_, ctx) => ctx.mode !== 'view' }
    ],
    rules: [
      { required: props.mode === 'add', message: t('user.validation.passwordRequired'), trigger: 'blur' },
      { min: 6, message: t('user.validation.passwordLength'), trigger: 'blur' },
    ],
  },
  {
    prop: 'confirmPassword',
    label: t('user.field.confirmPassword'),
    type: 'password',
    span: 12,
    dependencies: [
      { trigger: 'mode', show: (_, ctx) => ctx.mode !== 'view' }
    ],
    rules: [
      {
        required: props.mode === 'add',
        message: t('user.validation.confirmPasswordRequired'),
        trigger: 'blur',
      },
      {
        validator: (_rule, value, cb) => {
          if (value !== formData.password) {
            cb(new Error(t('user.validation.confirmMismatch')))
          } else {
            cb()
          }
        },
        trigger: ['blur', 'change'],
      },
    ],
  },
])

// password 字段变化时手动触发 confirmPassword 重校验（EP 2.11 不支持 relations prop）
watch(() => formData.password, () => {
  if (formData.confirmPassword) {
    formDrawerRef.value?.validateField('confirmPassword')
  }
})

const rules = computed(() => ({
  username: [
    { required: true, message: t('user.validation.usernameRequired'), trigger: 'blur' },
    { min: 3, max: 20, message: t('user.validation.usernameLength'), trigger: 'blur' },
  ],
  realName: [
    { required: true, message: t('user.validation.realNameRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('user.validation.realNameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: t('user.validation.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('user.validation.emailInvalid'), trigger: ['blur', 'change'] },
  ],
  phone: [
    { required: true, message: t('user.validation.phoneRequired'), trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: t('user.validation.phoneInvalid'), trigger: ['blur', 'change'] },
  ],
  roles: [
    { required: true, message: t('user.validation.rolesRequired'), trigger: 'change' },
    {
      validator: (_rule, value, cb) => {
        if (!Array.isArray(value) || value.length === 0) {
          cb(new Error(t('user.validation.rolesAtLeastOne')))
        } else {
          cb()
        }
      },
      trigger: 'change',
    },
  ],
  status: [{ required: true, message: t('user.validation.statusRequired'), trigger: 'change' }],
}))

const initForm = () => {
  const empty = {
    username: '',
    realName: '',
    email: '',
    phone: '',
    roles: [] as string[],
    status: 'active' as const,
    password: '',
    confirmPassword: '',
  }
  if ((props.mode === 'edit' || props.mode === 'view') && props.data) {
    const u = props.data
    Object.assign(formData, {
      username: u.username,
      realName: u.realName,
      email: u.email,
      phone: u.phone,
      roles: u.roles ?? [],
      status: u.status,
      password: '',
      confirmPassword: '',
    })
  } else {
    Object.assign(formData, empty)
  }
}

const handleSubmit = async (data: Record<string, unknown>) => {
  submitting.value = true
  try {
    const { confirmPassword: _omit, ...rest } = formData
    const payload = { ...rest, ...data } as UserCreateRequest
    // 编辑模式下若密码留空，则不修改密码
    if (props.mode === 'edit' && !payload.password) {
      delete payload.password
    }
    if (props.mode === 'add') {
      await createUser(payload)
      ElMessage.success(t('common.message.createSuccess'))
    } else {
      await updateUser(props.data!.id, payload)
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
