<template>
  <div class="topbar">
    <h3>{{ title }}</h3>
    <el-space>
      <el-button :loading="loading" @click="$emit('refresh')">{{ t('common.actions.refresh') }}</el-button>
      <el-button v-if="showFormat" @click="$emit('format')">{{ t('common.actions.format') }}</el-button>
      <el-button v-if="showPreview" @click="$emit('preview')">{{ t('common.actions.previewDiff') }}</el-button>
      <el-button v-if="showSave" type="primary" :loading="saving" :disabled="!canEdit" @click="$emit('save')">
        {{ t('common.actions.saveConfig') }}
      </el-button>
    </el-space>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

withDefaults(defineProps<{
  title: string
  loading?: boolean
  saving?: boolean
  canEdit?: boolean
  showFormat?: boolean
  showPreview?: boolean
  showSave?: boolean
}>(), {
  showPreview: true,
  showSave: true,
})

defineEmits<{
  refresh: []
  format: []
  preview: []
  save: []
}>()
</script>

<style scoped>
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.topbar h3 {
  margin: 0;
}
</style>
