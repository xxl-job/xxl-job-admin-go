<template>
  <div class="login-wrap">
    <!-- 左侧品牌区 -->
    <div class="brand-panel">
      <div class="brand-inner">
        <div class="brand-logo">
          <el-icon :size="40"><DataLine /></el-icon>
        </div>
        <h1 class="brand-title">XXL-JOB-Admin-Go</h1>
        <p class="brand-slogan">Go 实现的分布式任务调度中心</p>
        <ul class="brand-features">
          <li><el-icon><Connection /></el-icon> gRPC 长连接调度</li>
          <li><el-icon><Lock /></el-icon> MySQL 租约锁 · 无 Redis</li>
          <li><el-icon><Monitor /></el-icon> 实时日志 · 指标可观测</li>
        </ul>
      </div>
    </div>

    <!-- 右侧表单区 -->
    <div class="form-panel">
      <div class="dark-toggle" @click="themeStore.toggleDark()">
        <el-icon :size="20">
          <Moon v-if="themeStore.isDark" />
          <Sunny v-else />
        </el-icon>
      </div>

      <div class="form-card">
        <h2 class="form-title">{{ t('auth.title') }}</h2>
        <p class="form-sub">请输入账号密码登录调度中心</p>

        <el-form
          ref="ruleFormRef"
          :model="ruleForm"
          status-icon
          :rules="rules"
          label-position="top"
          size="large"
          @keyup.enter="submitForm(ruleFormRef)"
        >
          <el-form-item :label="t('auth.username')" prop="username">
            <el-input
              v-model="ruleForm.username"
              :placeholder="t('auth.username')"
              :prefix-icon="User"
            />
          </el-form-item>
          <el-form-item :label="t('auth.password')" prop="password">
            <el-input
              v-model="ruleForm.password"
              type="password"
              autocomplete="off"
              :placeholder="t('auth.password')"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="submitting"
              class="submit-btn"
              @click="submitForm(ruleFormRef)"
            >
              {{ t('auth.submit') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/app/stores/theme'
import { ElMessage, type FormInstance, type FormItemRule, type FormRules } from 'element-plus'
import { User, Lock, DataLine, Connection, Monitor } from '@element-plus/icons-vue'
import { authService } from '@/lib/auth/authService'
import { useUserStore } from '@/app/stores/user'
import { HttpError } from '@/lib/error/types'

const { t } = useI18n()
const ruleFormRef = ref<FormInstance>()
const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
const submitting = ref(false)

// 强类型校验器：禁止 any，类型由 FormItemRule['validator'] 推导
// trim 判空堵纯空格绕过；用户名 3-20、密码 ≥6 与 UserFormDrawer 创建规则对齐
const validateUsername: NonNullable<FormItemRule['validator']> = (
  _rule,
  value,
  callback,
) => {
  const v = String(value ?? '').trim()
  if (v === '') {
    callback(new Error(t('common.message.fieldRequired')))
  } else if (v.length < 3 || v.length > 20) {
    callback(new Error(t('auth.usernameLength')))
  } else {
    callback()
  }
}

const validatePassword: NonNullable<FormItemRule['validator']> = (
  _rule,
  value,
  callback,
) => {
  const v = String(value ?? '')
  if (v.trim() === '') {
    callback(new Error(t('common.message.fieldRequired')))
  } else if (v.length < 6) {
    callback(new Error(t('auth.passwordLength')))
  } else {
    callback()
  }
}

const ruleForm = reactive({
  username: '',
  password: '',
})

const rules = reactive<FormRules>({
  username: [{ validator: validateUsername, trigger: ['blur', 'change'] }],
  password: [{ validator: validatePassword, trigger: ['blur', 'change'] }],
})

const submitForm = async (formEl: FormInstance | undefined) => {
  if (!formEl) return
  try {
    await formEl.validate()
  } catch {
    return
  }
  submitting.value = true
  try {
    await authService.login(ruleForm)
    await userStore.loadProfile()
    router.push('/')
  } catch (e) {
    // auth 请求 _silent 抑制了拦截器全局提示，此处做领域内反馈：
    // HttpError 显示后端文案（如"用户名或密码错误"），其他异常走 i18n fallback
    ElMessage.error(e instanceof HttpError ? e.problem.title : t('auth.loginFailed'))
    console.debug('[login] failed', e)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  display: flex;
  min-height: 100vh;
  background: var(--el-bg-color-page);
}

/* 左侧品牌区 */
.brand-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44%;
  padding: 48px;
  background: linear-gradient(135deg, var(--el-color-primary) 0%, var(--el-color-primary-light-3) 100%);
  color: #fff;
}
.brand-inner {
  max-width: 380px;
}
.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  margin-bottom: 24px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(4px);
}
.brand-title {
  margin: 0 0 10px;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.brand-slogan {
  margin: 0 0 36px;
  font-size: 15px;
  opacity: 0.92;
}
.brand-features {
  margin: 0;
  padding: 0;
  list-style: none;
}
.brand-features li {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  font-size: 14px;
  opacity: 0.95;
}

/* 右侧表单区 */
.form-panel {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  padding: 48px 24px;
}
.dark-toggle {
  position: absolute;
  top: 24px;
  right: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  cursor: pointer;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  transition: background 0.2s;
}
.dark-toggle:hover {
  background: var(--el-fill-color);
}
.form-card {
  width: 100%;
  max-width: 380px;
  padding: 40px 36px 32px;
  background: var(--el-bg-color);
  border-radius: 14px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.06);
}
.form-title {
  margin: 0 0 6px;
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.form-sub {
  margin: 0 0 28px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
.submit-btn {
  width: 100%;
}

/* 响应式：窄屏隐藏品牌区 */
@media (max-width: 900px) {
  .brand-panel {
    display: none;
  }
}
</style>
