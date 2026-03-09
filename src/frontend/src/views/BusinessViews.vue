<template>
  <div class="business-page">
    <h3>{{ t('business.title') }}</h3>
    <p>{{ t('business.description') }}</p>
    <el-card shadow="never">
      <el-space wrap>
        <el-tag type="primary">{{ t('business.tags.config') }}</el-tag>
        <el-tag type="success">{{ t('business.tags.agents') }}</el-tag>
        <el-tag type="warning">{{ t('business.tags.skills') }}</el-tag>
        <el-tag type="info">{{ t('business.tags.backups') }}</el-tag>
      </el-space>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
<style scoped>
.business-page { display: grid; gap: 12px; }
</style>
