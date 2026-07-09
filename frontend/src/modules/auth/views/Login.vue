<template>
  <el-row
    class="login"
    justify="center"
    align="middle"
  >
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>{{ t('auth.title') }}</span>
          <div
            class="dark-icon"
            @click="themeStore.toggleDark()"
          >
            <el-icon>
              <Moon v-if="themeStore.isDark" />
              <Sunny v-else />
            </el-icon>
          </div>
        </div>
      </template>
      <div>
        <el-form
          ref="ruleFormRef"
          :model="ruleForm"
          status-icon
          :rules="rules"
          label-width="60px"
          class="demo-ruleForm"
        >
          <el-form-item
            :label="t('auth.username')"
            prop="username"
          >
            <el-input v-model="ruleForm.username" />
          </el-form-item>
          <el-form-item
            :label="t('auth.password')"
            prop="password"
          >
            <el-input
              v-model="ruleForm.password"
              type="password"
              autocomplete="off"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="submitting"
              @click="submitForm(ruleFormRef)"
            >
              {{ t('auth.submit') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </el-row>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from '@/app/stores/theme'
import { ElMessage, type FormInstance, type FormItemRule, type FormRules } from 'element-plus'
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
.login {
  position: absolute;
  top: 20%;
  bottom: 60%;
  width: 100%;
}

.box-card {
  width: 450px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dark-icon {
  font-size: 20px;
  cursor: pointer;
}

.text {
  font-size: 14px;
}

.item {
  margin-bottom: 18px;
}
</style>
