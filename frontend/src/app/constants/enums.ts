/**
 * 业务枚举映射集中常量。
 *
 * 用途：消除 StatusTag 业务映射、通知优先级、角色类型等在多组件重复定义。
 * 类型策略：`type` 字段直接对齐 Element Plus 的 Tag/Alert `type` 取值。
 */

import { t } from '@/lib/i18n'

// Element Plus 组件 type 取值（el-tag / el-alert 共用）
export type EpType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

// 状态映射条目类型（与 StatusTag 内部 StatusMap 一致）
export type StatusMapEntry = Record<
  string,
  { type: EpType; text: string }
>

/** 通用启用/禁用状态 */
export const COMMON_STATUS_MAP: StatusMapEntry = {
  active: { type: 'success', text: '启用' },
  inactive: { type: 'info', text: '禁用' },
}

/** 权限模块（permission List 用）。text 存 i18n key，由 StatusTag 经 t() 翻译。 */
export const MODULE_STATUS_MAP: StatusMapEntry = {
  system: { type: 'primary', text: 'permission.option.moduleSystem' },
  user: { type: 'info', text: 'permission.option.moduleUser' },
  role: { type: 'warning', text: 'permission.option.moduleRole' },
  permission: { type: 'success', text: 'permission.option.modulePermission' },
  dict: { type: 'danger', text: 'permission.option.moduleDict' },
  config: { type: 'info', text: 'permission.option.moduleConfig' },
}

/** 通知优先级 i18n key（由 priorityLabel 经 t() 翻译） */
export const PRIORITY_LABEL: Record<string, string> = {
  high: 'notice.priority.high',
  medium: 'notice.priority.medium',
  low: 'notice.priority.low',
}

/** 通知优先级 → el-tag type 映射 */
export const PRIORITY_TAG_TYPE: Record<string, EpType> = {
  high: 'danger',
  medium: 'warning',
  low: 'info',
}

/** 通知优先级 → el-alert type 映射 */
export const PRIORITY_ALERT_TYPE: Record<string, 'error' | 'warning' | 'info'> = {
  high: 'error',
  medium: 'warning',
  low: 'info',
}

/** 通知类型 i18n key（由 noticeTypeLabel 经 t() 翻译） */
export const NOTICE_TYPE_LABEL: Record<string, string> = {
  announcement: 'notice.option.typeAnnouncement',
  notice: 'notice.option.typeNotice',
  todo: 'notice.option.typeTodo',
}

/** 工具：根据优先级 code 取 label（i18n key 经 t() 翻译） */
export function priorityLabel(code?: string): string {
  return t(PRIORITY_LABEL[code || 'low'])
}

/** 工具：根据优先级 code 取 el-tag type */
export function priorityTagType(code?: string): EpType {
  return PRIORITY_TAG_TYPE[code || 'low'] || 'info'
}

/** 工具：根据优先级 code 取 el-alert type */
export function priorityAlertType(
  code?: string
): 'error' | 'warning' | 'info' {
  return PRIORITY_ALERT_TYPE[code || 'low'] || 'info'
}

/** 工具：根据通知 type 取 label（i18n key 经 t() 翻译） */
export function noticeTypeLabel(code?: string): string {
  return t(NOTICE_TYPE_LABEL[code || 'announcement'])
}

/** 通用启用/禁用 select 选项（FormDrawer 字段配置直接复用）。
 * 用函数包装以在每次调用时求值 t()，保持 locale 切换响应式。 */
export const getCommonStatusOptions = () => [
  { label: t('common.status.enable'), value: 'active' },
  { label: t('common.status.disable'), value: 'inactive' },
]

/** 通知优先级 select 选项 */
export const PRIORITY_OPTIONS = [
  { label: '紧急', value: 'high' },
  { label: '重要', value: 'medium' },
  { label: '普通', value: 'low' },
] as const

/** 通知类型 select 选项 */
export const NOTICE_TYPE_OPTIONS = [
  { label: '公告', value: 'announcement' },
  { label: '通知', value: 'notice' },
  { label: '待办', value: 'todo' },
] as const
