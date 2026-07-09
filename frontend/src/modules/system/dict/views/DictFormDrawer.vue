<template>
  <el-drawer
    v-model="visible"
    :title="title"
    size="50%"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      class="dict-form"
    >
      <!-- 分类表单 -->
      <template v-if="formType === 'category'">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.categoryName')"
              prop="name"
            >
              <el-input
                v-model="form.name"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.categoryName') })"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.categoryCode')"
              prop="code"
            >
              <el-input
                v-model="form.code"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.categoryCode') })"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item
              :label="t('common.column.description')"
              prop="description"
            >
              <el-input
                v-model="form.description"
                :placeholder="t('common.placeholder.input', { label: t('common.column.description') })"
                type="textarea"
                :rows="3"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.status')"
              prop="status"
            >
              <el-select
                v-model="form.status"
                :placeholder="t('common.placeholder.select', { label: t('dict.field.status') })"
              >
                <el-option
                  v-for="opt in getCommonStatusOptions()"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- 字典表单 -->
      <template v-else-if="formType === 'dict'">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.dictName')"
              prop="name"
            >
              <el-input
                v-model="form.name"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.dictName') })"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.dictCode')"
              prop="code"
            >
              <el-input
                v-model="form.code"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.dictCode') })"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item
              :label="t('common.column.description')"
              prop="description"
            >
              <el-input
                v-model="form.description"
                :placeholder="t('common.placeholder.input', { label: t('common.column.description') })"
                type="textarea"
                :rows="3"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.status')"
              prop="status"
            >
              <el-select
                v-model="form.status"
                :placeholder="t('common.placeholder.select', { label: t('dict.field.status') })"
              >
                <el-option
                  v-for="opt in getCommonStatusOptions()"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <!-- 字典项表单 -->
      <template v-else>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.itemName')"
              prop="name"
            >
              <el-input
                v-model="form.name"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.itemName') })"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.itemCode')"
              prop="code"
            >
              <el-input
                v-model="form.code"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.itemCode') })"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.dictValue')"
              prop="value"
            >
              <el-input
                v-model="form.value"
                :placeholder="t('common.placeholder.input', { label: t('dict.field.dictValue') })"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.sort')"
              prop="sort"
            >
              <el-input-number
                v-model="form.sort"
                :min="0"
                :max="999"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item
              :label="t('dict.field.status')"
              prop="status"
            >
              <el-select
                v-model="form.status"
                :placeholder="t('common.placeholder.select', { label: t('dict.field.status') })"
              >
                <el-option
                  v-for="opt in getCommonStatusOptions()"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </template>

      <el-form-item>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ mode === 'add' ? t('common.action.create') : t('common.action.save') }}
        </el-button>
        <el-button @click="visible = false">
          {{ t('common.action.cancel') }}
        </el-button>
      </el-form-item>
    </el-form>
  </el-drawer>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { DictFormState, DictTreeNode } from './hooks/useDictTree'
import { t } from '@/lib/i18n'
import { getCommonStatusOptions } from '@/app/constants/enums'

const props = defineProps<{
  modelValue: boolean
  mode: 'add' | 'edit'
  /**
   * 新增模式下：父节点（level 0=虚拟根 / 1=分类 / 2=字典）
   * 编辑模式下：被编辑节点（level 1/2/3）
   */
  parentNode: DictTreeNode | { level: 0 } | null
  selectedNode: DictTreeNode | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit'): void
}>()

// form 作为 model：子组件 v-model 双向绑定字段，符合 Vue 单向数据流
const form = defineModel<DictFormState>('form', { required: true })

const formRef = ref<FormInstance>()
const submitting = ref(false)

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const formType = computed<'category' | 'dict' | 'item'>(() => {
  if (props.mode === 'add') {
    const lvl = props.parentNode?.level ?? 0
    if (lvl === 0) return 'category'
    if (lvl === 1) return 'dict'
    return 'item'
  }
  const lvl = props.selectedNode?.level ?? 1
  if (lvl === 1) return 'category'
  if (lvl === 2) return 'dict'
  return 'item'
})

const title = computed(() => {
  const action = props.mode === 'add' ? t('common.action.create') : t('common.action.edit')
  const typeMap = { category: t('dict.level.1'), dict: t('dict.level.2'), item: t('dict.level.3') }
  return `${action}${typeMap[formType.value]}`
})

const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('dict.validation.nameRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('dict.validation.nameLength'), trigger: 'blur' },
  ],
  code: [
    { required: true, message: t('dict.validation.codeRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: t('dict.validation.codeLength'), trigger: 'blur' },
  ],
  description: [
    { max: 200, message: t('dict.validation.descriptionMax'), trigger: 'blur' },
  ],
  status: [{ required: true, message: t('dict.validation.statusRequired'), trigger: 'change' }],
  value: [
    { required: true, message: t('dict.validation.valueRequired'), trigger: 'blur' },
    { min: 1, max: 50, message: t('dict.validation.valueLength'), trigger: 'blur' },
  ],
  sort: [{ required: true, message: t('dict.validation.sortRequired'), trigger: 'blur' }],
}))

// 切换为不同节点时，清理上一次的校验状态
watch(
  () => props.modelValue,
  (v) => {
    if (v) {
      formRef.value?.clearValidate()
    }
  }
)

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  submitting.value = true
  emit('submit')
  submitting.value = false
}

const handleClosed = () => {
  formRef.value?.clearValidate()
}
</script>

<style scoped>
.dict-form {
  padding: 0 20px;
}
</style>
