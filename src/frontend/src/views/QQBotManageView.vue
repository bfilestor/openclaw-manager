<template>
  <div class="qqbot-page">
    <OpenclawSaveActions
      title="QQBot 管理"
      :loading="loading"
      :saving="saving"
      :can-edit="canEdit"
      @refresh="loadConfig"
      @preview="previewDiff"
      @save="saveConfig"
    />

    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>已接入 QQBot 列表</span>
          <el-tag type="info">共 {{ bots.length }} 个</el-tag>
        </div>
      </template>

      <el-empty v-if="!loading && bots.length === 0" description="尚未接入任何 QQBot" />

      <el-table v-else :data="bots" row-key="key" style="width: 100%">
        <el-table-column prop="name" label="Bot 名称" min-width="140" />
        <el-table-column label="AppID" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.appId" :disabled="!canEdit" placeholder="请输入 AppID" />
          </template>
        </el-table-column>
        <el-table-column label="AppSecret" min-width="260">
          <template #default="{ row }">
            <el-input v-model="row.clientSecret" :disabled="!canEdit" placeholder="请输入 AppSecret" show-password />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="danger" link :disabled="!canEdit" @click="removeBot(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-alert
        v-if="bots.length === 0"
        class="tips"
        type="info"
        :closable="false"
        show-icon
        title="首次接入提示：输入主 QQBot 的 AppID 和 AppSecret 后保存，然后执行下方 3 行命令完成第一次接入。"
      />

      <div v-if="bots.length === 0" class="first-access">
        <el-form label-position="top">
          <el-form-item label="主 QQBot AppID">
            <el-input v-model="firstBot.appId" :disabled="!canEdit" placeholder="请输入第一个 bot 的 AppID" />
          </el-form-item>
          <el-form-item label="主 QQBot AppSecret">
            <el-input
              v-model="firstBot.clientSecret"
              :disabled="!canEdit"
              show-password
              placeholder="请输入第一个 bot 的 AppSecret"
            />
          </el-form-item>
        </el-form>

        <div class="cmd-box">
          <div class="cmd-title">第一次接入命令（保存后执行）</div>
          <pre>cd /home/mixi/.openclaw/projects/manager
scripts/build.sh
openclaw gateway restart</pre>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <template #header>添加新 QQBot（支持一次添加多个）</template>

      <el-alert
        type="warning"
        show-icon
        :closable="false"
        title="如已有主 QQBot，新增会写入 qqbot.accounts。bot 名称仅允许字母和数字（例如 qqbot2）。"
      />

      <div class="add-list">
        <div v-for="(item, idx) in addRows" :key="item.id" class="add-row">
          <el-row :gutter="10">
            <el-col :xs="24" :md="6">
              <el-input
                v-model="item.name"
                :disabled="!canEdit"
                placeholder="bot名称，如 qqbot2"
              />
            </el-col>
            <el-col :xs="24" :md="8">
              <el-input v-model="item.appId" :disabled="!canEdit" placeholder="AppID" />
            </el-col>
            <el-col :xs="24" :md="8">
              <el-input v-model="item.clientSecret" :disabled="!canEdit" placeholder="AppSecret" show-password />
            </el-col>
            <el-col :xs="24" :md="2" class="row-op">
              <el-button type="danger" :disabled="!canEdit" @click="removeAddRow(idx)">删除</el-button>
            </el-col>
          </el-row>
        </div>
      </div>

      <el-space>
        <el-button :disabled="!canEdit" @click="addRow">+ 添加一行</el-button>
        <el-button type="primary" :disabled="!canEdit" @click="appendRowsToConfig">加入配置草稿</el-button>
      </el-space>
    </el-card>

    <el-dialog v-model="diffDialogVisible" width="1080px" title="openclaw.json 变更预览">
      <DiffViewer :from-text="diffFromText" :to-text="diffToText" :height="460" />
      <template #footer>
        <el-button @click="diffDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import DiffViewer from '../components/DiffViewer.vue'
import OpenclawSaveActions from '../components/OpenclawSaveActions.vue'
import { buildOpenclawDiff, getOpenclawConfig, saveOpenclawConfig } from '../services/openclawConfig'

type BotRow = {
  key: string
  name: string
  isPrimary: boolean
  appId: string
  clientSecret: string
}

type AddRow = {
  id: string
  name: string
  appId: string
  clientSecret: string
}

const auth = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')

const rawConfig = ref<any>({})
const originalConfigText = ref('{}')
const bots = ref<BotRow[]>([])

const firstBot = ref({ appId: '', clientSecret: '' })
const addRows = ref<AddRow[]>([{ id: crypto.randomUUID(), name: '', appId: '', clientSecret: '' }])

const diffDialogVisible = ref(false)
const diffFromText = ref('')
const diffToText = ref('')

const canEdit = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function listBotsFromConfig(cfg: any): BotRow[] {
  const out: BotRow[] = []
  const qqbot = cfg?.channels?.qqbot
  if (!qqbot || typeof qqbot !== 'object') return out

  out.push({
    key: 'primary',
    name: 'qqbot',
    isPrimary: true,
    appId: String(qqbot.appId || ''),
    clientSecret: String(qqbot.clientSecret || ''),
  })

  const accounts = qqbot.accounts
  if (accounts && typeof accounts === 'object') {
    Object.keys(accounts).forEach((name) => {
      const item = accounts[name] || {}
      out.push({
        key: `acc:${name}`,
        name,
        isPrimary: false,
        appId: String(item.appId || ''),
        clientSecret: String(item.clientSecret || ''),
      })
    })
  }
  return out
}

function ensureQQBotRoot(cfg: any) {
  if (!cfg.channels || typeof cfg.channels !== 'object') cfg.channels = {}
  if (!cfg.channels.qqbot || typeof cfg.channels.qqbot !== 'object') {
    cfg.channels.qqbot = {
      enabled: true,
      allowFrom: ['*'],
      appId: '',
      clientSecret: '',
    }
  }
}

async function loadConfig() {
  loading.value = true
  errorMessage.value = ''
  try {
    const payload = await getOpenclawConfig()
    const text = String(payload.content || '{}')
    originalConfigText.value = text
    rawConfig.value = JSON.parse(text)
    bots.value = listBotsFromConfig(rawConfig.value)
  } catch (err) {
    errorMessage.value = parseError(err, '加载配置失败')
  } finally {
    loading.value = false
  }
}

function removeBot(row: BotRow) {
  if (row.isPrimary) {
    const qqbot = rawConfig.value?.channels?.qqbot
    if (qqbot && typeof qqbot === 'object') {
      delete rawConfig.value.channels.qqbot
    }
  } else {
    const accounts = rawConfig.value?.channels?.qqbot?.accounts
    if (accounts && typeof accounts === 'object') {
      delete accounts[row.name]
    }
  }
  bots.value = listBotsFromConfig(rawConfig.value)
}

function addRow() {
  addRows.value.push({ id: crypto.randomUUID(), name: '', appId: '', clientSecret: '' })
}

function removeAddRow(idx: number) {
  addRows.value.splice(idx, 1)
  if (addRows.value.length === 0) addRow()
}

function ensureValidBotName(name: string): boolean {
  return /^[A-Za-z0-9]+$/.test(name)
}

function appendRowsToConfig() {
  if (!canEdit.value) return

  if (bots.value.length === 0) {
    if (!firstBot.value.appId.trim() || !firstBot.value.clientSecret.trim()) {
      ElMessage.warning('请先输入主 QQBot 的 AppID 和 AppSecret')
      return
    }
    ensureQQBotRoot(rawConfig.value)
    rawConfig.value.channels.qqbot.enabled = true
    rawConfig.value.channels.qqbot.allowFrom = Array.isArray(rawConfig.value.channels.qqbot.allowFrom)
      ? rawConfig.value.channels.qqbot.allowFrom
      : ['*']
    rawConfig.value.channels.qqbot.appId = firstBot.value.appId.trim()
    rawConfig.value.channels.qqbot.clientSecret = firstBot.value.clientSecret.trim()
    bots.value = listBotsFromConfig(rawConfig.value)
    ElMessage.success('主 QQBot 已加入配置草稿')
    return
  }

  ensureQQBotRoot(rawConfig.value)
  const qqbot = rawConfig.value.channels.qqbot
  if (!qqbot.accounts || typeof qqbot.accounts !== 'object') qqbot.accounts = {}

  let added = 0
  for (const row of addRows.value) {
    const name = row.name.trim()
    const appId = row.appId.trim()
    const clientSecret = row.clientSecret.trim()
    if (!name && !appId && !clientSecret) continue

    if (!ensureValidBotName(name)) {
      ElMessage.error(`Bot 名称不合法：${name || '(空)'}。仅允许字母和数字`)
      return
    }
    if (!appId || !clientSecret) {
      ElMessage.error(`Bot ${name} 缺少 AppID 或 AppSecret`)
      return
    }

    qqbot.accounts[name] = {
      enabled: true,
      appId,
      clientSecret,
      allowFrom: ['*'],
    }
    added += 1
  }

  if (added === 0) {
    ElMessage.warning('没有可添加的数据，请填写至少一行')
    return
  }

  bots.value = listBotsFromConfig(rawConfig.value)
  addRows.value = [{ id: crypto.randomUUID(), name: '', appId: '', clientSecret: '' }]
  ElMessage.success(`已加入 ${added} 个 QQBot 到配置草稿`)
}

function buildNormalizedConfigText(): string {
  // 同步页面已编辑的 bot 列表到配置对象
  if (bots.value.length > 0) {
    ensureQQBotRoot(rawConfig.value)
    const qqbot = rawConfig.value.channels.qqbot
    const primary = bots.value.find((x) => x.isPrimary)
    if (primary) {
      qqbot.enabled = true
      qqbot.allowFrom = Array.isArray(qqbot.allowFrom) ? qqbot.allowFrom : ['*']
      qqbot.appId = primary.appId.trim()
      qqbot.clientSecret = primary.clientSecret.trim()
    }

    const accounts: Record<string, any> = {}
    bots.value.filter((x) => !x.isPrimary).forEach((row) => {
      accounts[row.name] = {
        enabled: true,
        appId: row.appId.trim(),
        clientSecret: row.clientSecret.trim(),
        allowFrom: ['*'],
      }
    })
    if (Object.keys(accounts).length > 0) qqbot.accounts = accounts
    else delete qqbot.accounts
  }

  return JSON.stringify(rawConfig.value, null, 2)
}

function previewDiff() {
  try {
    const diff = buildOpenclawDiff(originalConfigText.value, buildNormalizedConfigText())
    diffFromText.value = diff.fromText
    diffToText.value = diff.toText
    diffDialogVisible.value = true
  } catch {
    ElMessage.error('生成变更预览失败，请检查输入内容')
  }
}

async function saveConfig() {
  if (!canEdit.value) {
    ElMessage.warning('当前角色无编辑权限')
    return
  }

  let normalized = ''
  try {
    normalized = buildNormalizedConfigText()
  } catch {
    ElMessage.error('配置序列化失败，请检查输入内容')
    return
  }

  try {
    await ElMessageBox.confirm('确认保存 QQBot 配置变更？保存前将按 revision 机制记录版本。', '保存确认', { type: 'warning' })
  } catch {
    return
  }

  saving.value = true
  try {
    await saveOpenclawConfig(normalized)
    ElMessage.success('QQBot 配置保存成功')
    await loadConfig()
  } catch (err) {
    ElMessage.error(parseError(err, '保存配置失败'))
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.qqbot-page {
  display: grid;
  gap: 12px;
}
.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.tips {
  margin-top: 10px;
}
.first-access {
  margin-top: 12px;
}
.cmd-box {
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  padding: 10px;
  background: #fafafa;
}
.cmd-title {
  font-weight: 600;
  margin-bottom: 6px;
}
.cmd-box pre {
  margin: 0;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
}
.add-list {
  margin: 10px 0;
  display: grid;
  gap: 8px;
}
.add-row {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 8px;
}
.row-op {
  display: flex;
  justify-content: flex-end;
}
</style>
