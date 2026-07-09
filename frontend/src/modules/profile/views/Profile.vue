<template>
  <div class="user-portrait">
    <el-card
      shadow="never"
      class="portrait-card"
    >
      <template #header>
        <div class="card-header">
          <span>{{ t('profile.title') }}</span>
          <el-button
            type="primary"
            size="small"
            @click="handleEdit"
          >
            {{ t('profile.editProfile') }}
          </el-button>
        </div>
      </template>

      <div class="portrait-content">
        <!-- 用户头像 -->
        <div class="avatar-section">
          <el-avatar
            :size="120"
            :src="userInfo.avatar"
          >
            <el-icon><User /></el-icon>
          </el-avatar>
          <div class="username">
            {{ userInfo.name }}
          </div>
          <div class="role">
            {{ userInfo.role }}
          </div>
        </div>

        <!-- 基本信息 -->
        <el-divider content-position="left">
          {{ t('profile.basicInfo') }}
        </el-divider>
        <el-descriptions
          :column="2"
          border
        >
          <el-descriptions-item :label="t('profile.field.username')">
            {{
              userInfo.username
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.email')">
            {{
              userInfo.email
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.phone')">
            {{
              userInfo.phone
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.gender')">
            {{
              userInfo.gender
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.department')">
            {{
              userInfo.department
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.position')">
            {{
              userInfo.position
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.joinDate')">
            {{
              userInfo.joinDate
            }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('profile.field.lastLogin')">
            {{
              userInfo.lastLogin
            }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 个人设置 -->
        <el-divider content-position="left">
          {{ t('profile.personalSettings') }}
        </el-divider>
        <el-card
          shadow="never"
          class="settings-card"
        >
          <el-form
            :model="settings"
            label-width="100px"
          >
            <el-form-item :label="t('profile.field.language')">
              <el-select
                v-model="settings.language"
                style="width: 100%"
              >
                <el-option
                  :label="t('profile.option.chinese')"
                  value="zh-CN"
                />
                <el-option
                  :label="t('profile.option.english')"
                  value="en-US"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="t('profile.field.timezone')">
              <el-select
                v-model="settings.timezone"
                style="width: 100%"
              >
                <el-option
                  label="UTC+8"
                  value="Asia/Shanghai"
                />
                <el-option
                  label="UTC+0"
                  value="Europe/London"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="t('profile.field.notification')">
              <el-switch v-model="settings.notification" />
            </el-form-item>
            <el-form-item :label="t('profile.field.emailReminder')">
              <el-switch v-model="settings.emailReminder" />
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </el-card>

    <!-- 编辑资料对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="t('profile.editProfile')"
      width="50%"
    >
      <el-form
        :model="userInfo"
        label-width="100px"
      >
        <el-form-item :label="t('profile.field.name')">
          <el-input v-model="userInfo.name" />
        </el-form-item>
        <el-form-item :label="t('profile.field.email')">
          <el-input v-model="userInfo.email" />
        </el-form-item>
        <el-form-item :label="t('profile.field.phone')">
          <el-input v-model="userInfo.phone" />
        </el-form-item>
        <el-form-item :label="t('profile.field.gender')">
          <el-radio-group v-model="userInfo.gender">
            <el-radio value="男">
              男
            </el-radio>
            <el-radio value="女">
              女
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('profile.field.department')">
          <el-input v-model="userInfo.department" />
        </el-form-item>
        <el-form-item :label="t('profile.field.position')">
          <el-input v-model="userInfo.position" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.action.cancel') }}</el-button>
          <el-button
            type="primary"
            @click="handleSave"
          >{{ t('common.action.save') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive } from 'vue'
import { User } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { t } from '@/lib/i18n'

// 用户信息
const userInfo = reactive({
  name: '张三',
  username: 'zhangsan',
  role: '管理员',
  email: 'zhangsan@example.com',
  phone: '13800138000',
  gender: '男',
  department: '技术部',
  position: '前端工程师',
  joinDate: '2023-01-15',
  lastLogin: '2023-10-01 10:30',
  avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
})

// 个人设置
const settings = reactive({
  language: 'zh-CN',
  timezone: 'Asia/Shanghai',
  notification: true,
  emailReminder: false
})

// 对话框可见性
const dialogVisible = ref(false)

// 编辑资料
const handleEdit = () => {
  dialogVisible.value = true
}

// 保存资料
const handleSave = () => {
  dialogVisible.value = false
  ElMessage.success(t('profile.saveSuccess'))
}
</script>

<style lang="scss" scoped>
.portrait-card {
  margin-bottom: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 600;
    font-size: 16px;
  }

  .portrait-content {
    .avatar-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin-bottom: 30px;

      .username {
        margin-top: 10px;
        font-size: 18px;
        font-weight: 600;
        color: #303133;
      }

      .role {
        margin-top: 5px;
        font-size: 14px;
        color: #606266;
      }
    }

    .settings-card {
      margin-top: 20px;
    }
  }
}

.dialog-footer {
  text-align: right;
}
</style>
