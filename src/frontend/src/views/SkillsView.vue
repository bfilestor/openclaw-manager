<template>
  <div class="skills-page">
    <div class="topbar">
      <h3>Skills</h3>
      <el-button :loading="loading" @click="loadSkills">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="stats-row">
      <el-col :xs="24" :sm="12">
        <el-card shadow="never">已安装技能: {{ skills.length }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="never">总占用空间: {{ formatBytes(totalBytes) }}</el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="skills" row-key="name" style="width: 100%">
        <el-table-column prop="name" label="Skill 名称" min-width="220" />
        <el-table-column label="作用域" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.scope }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="元信息" width="120">
          <template #default="{ row }">
            <el-tag :type="row.has_meta ? 'success' : 'warning'">
              {{ row.has_meta ? '完整' : '缺失' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="160">
          <template #default="{ row }">{{ formatBytes(row.size_bytes) }}</template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && skills.length === 0"
        description="当前系统还没有已安装的 Skill"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

type SkillItem = {
  name: string
  scope: string
  agent_id?: string
  size_bytes: number
  has_meta: boolean
}

const loading = ref(false)
const errorMessage = ref('')
const skills = ref<SkillItem[]>([])
const totalBytes = computed(() => skills.value.reduce((sum, it) => sum + (it.size_bytes || 0), 0))

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / Math.pow(1024, exp)
  return `${value.toFixed(value >= 10 || exp === 0 ? 0 : 1)} ${units[exp]}`
}

async function loadSkills() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get('/api/v1/skills', { params: { scope: 'global' } })
    skills.value = Array.isArray(data?.skills) ? data.skills : []
  } catch {
    skills.value = []
    errorMessage.value = '加载 Skill 列表失败，请检查服务状态后重试'
    ElMessage.error('加载 Skill 列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadSkills)
</script>

<style scoped>
.skills-page {
  display: grid;
  gap: 12px;
}
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.topbar h3 {
  margin: 0;
}
.stats-row {
  margin: 0;
}
</style>
