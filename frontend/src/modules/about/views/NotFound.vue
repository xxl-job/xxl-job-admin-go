<template>
  <div class="not-found">
    <el-result
      icon="warning"
      title="404"
      :sub-title="t('common.notFound.subTitle')"
    >
      <template #extra>
        <el-button
          type="primary"
          @click="goHome"
        >
          {{ t('common.notFound.backHome') }}
        </el-button>
      </template>
    </el-result>
  </div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router'
import { t } from '@/lib/i18n'

const router = useRouter()
const goHome = () => router.push('/')
</script>

<style scoped>
.not-found {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}
</style>
