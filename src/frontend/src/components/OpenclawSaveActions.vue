<template>
  <div class="topbar">
    <h3>{{ title }}</h3>
    <el-space>
      <el-button :loading="loading" @click="$emit('refresh')">刷新</el-button>
      <el-button v-if="showFormat" @click="$emit('format')">格式化</el-button>
      <el-button @click="$emit('preview')">保存前预览 Diff</el-button>
      <el-button type="primary" :loading="saving" :disabled="!canEdit" @click="$emit('save')">
        保存配置
      </el-button>
    </el-space>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  loading?: boolean
  saving?: boolean
  canEdit?: boolean
  showFormat?: boolean
}>()

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
