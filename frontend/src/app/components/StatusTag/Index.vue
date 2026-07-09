<template>
  <el-tag
    :type="tagType"
    :size="size"
    :effect="effect"
  >
    {{ displayText }}
  </el-tag>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { t } from '@/lib/i18n'

interface StatusMap {
  [key: string]: { type: string; text: string }
}

const props = withDefaults(
  defineProps<{
    /** 状态值 */
    status?: string | number | boolean
    /** 自定义映射 */
    statusMap?: StatusMap
    /** 自定义显示文本 */
    text?: string
    /** 标签大小 */
    size?: 'large' | 'default' | 'small'
    /** 标签效果 */
    effect?: 'dark' | 'light' | 'plain'
  }>(),
  {
    status: '',
    statusMap: () => ({}),
    text: '',
    size: 'small',
    effect: 'light',
  }
)

// 默认状态映射：text 存 i18n key，由 displayText 经 t() 翻译（locale 切换响应式）
const defaultStatusMap: StatusMap = {
  // 启用/禁用
  active: { type: 'success', text: 'common.status.enable' },
  inactive: { type: 'info', text: 'common.status.disable' },
  // 布尔值
  true: { type: 'success', text: 'common.status.yes' },
  false: { type: 'info', text: 'common.status.no' },
  // 数字状态
  1: { type: 'success', text: 'common.status.enable' },
  0: { type: 'info', text: 'common.status.disable' },
  // 通用状态
  success: { type: 'success', text: 'common.status.success' },
  failed: { type: 'danger', text: 'common.status.failed' },
  error: { type: 'danger', text: 'common.status.error' },
  pending: { type: 'warning', text: 'common.status.pending' },
  processing: { type: 'primary', text: 'common.status.processing' },
  completed: { type: 'success', text: 'common.status.completed' },
  cancelled: { type: 'info', text: 'common.status.cancelled' },
  deleted: { type: 'danger', text: 'common.status.deleted' },
}

const mergedMap = computed<StatusMap>(() => ({
  ...defaultStatusMap,
  ...props.statusMap,
}))

const tagType = computed<string>(() => {
  const key = String(props.status ?? '')
  return mergedMap.value[key]?.type || 'info'
})

const displayText = computed<string>(() => {
  if (props.text) return props.text
  const key = String(props.status ?? '')
  const entry = mergedMap.value[key]
  // entry.text 可能是 i18n key（defaultStatusMap）或调用方传入的字面量（自定义 statusMap）
  // 字面量经 t() 缺失 key 会回退为原字符串，向后兼容
  return entry ? t(entry.text) : String(props.status ?? '')
})
</script>
