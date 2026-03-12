<template>
  <div class="shell-page">
    <div class="topbar">
      <h3>{{ t('shell.title') }}</h3>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="9">
        <el-card shadow="never" class="left-card">
          <template #header>
            <div class="panel-title">{{ t('shell.stashTitle') }}</div>
          </template>

          <el-input
            v-model="stashInput"
            type="textarea"
            :rows="8"
            :placeholder="t('shell.stashPlaceholder')"
          />

          <el-space class="stash-toolbar" wrap>
            <el-button @click="appendToStash">{{ t('shell.addToStash') }}</el-button>
            <el-button type="danger" :disabled="stashCommands.length === 0" @click="clearStash">{{ t('shell.clearStash') }}</el-button>
          </el-space>

          <el-table :data="stashCommands" row-key="id" style="width: 100%" height="300">
            <el-table-column prop="command" :label="t('shell.stashedCommands')" min-width="280" show-overflow-tooltip />
            <el-table-column :label="t('shell.columns.actions')" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="success" link @click="fillCommand(row.command)">{{ t('shell.fill') }}</el-button>
                <el-button type="danger" link @click="removeStash(row.id)">{{ t('common.actions.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="15">
        <el-card shadow="never" class="right-card">
          <template #header>
            <div class="panel-title">{{ t('shell.execAndLogs') }}</div>
          </template>

          <el-alert
            class="security-tip"
            type="warning"
            show-icon
            :closable="false"
            :title="t('shell.securityTip')"
          />

          <div class="command-row">
            <el-input
              v-model="commandInput"
              :placeholder="t('shell.commandPlaceholder')"
              @keyup.enter="executeCommand"
            />
            <el-button type="success" :loading="executing" :disabled="!canExecute" @click="executeCommand">
              {{ t('shell.execute') }}
            </el-button>
            <el-button :disabled="!logText" @click="clearLog">{{ t('shell.clearLogs') }}</el-button>
          </div>

          <pre ref="logBoxRef" class="log-box">{{ logText || t('shell.noLogs') }}</pre>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { executeShellCommand } from '../services/shell'

type StashCommand = {
  id: string
  command: string
}

const STORAGE_KEY = 'openclaw_manager_shell_stash_v1'
const auth = useAuthStore()
const { t } = useI18n()

const stashInput = ref('')
const stashCommands = ref<StashCommand[]>([])
const commandInput = ref('')
const executing = ref(false)
const logText = ref('')
const logBoxRef = ref<HTMLElement | null>(null)

const canExecute = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

function isOpenclawCommand(command: string): boolean {
  return /^openclaw(?:\s|$)/.test(command.trim())
}

function parseLines(raw: string): string[] {
  return raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

function newRowID(): string {
  const c = (globalThis as any)?.crypto
  if (c && typeof c.randomUUID === 'function') return c.randomUUID()
  return `cmd_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

function appendLogLine(line: string) {
  const text = String(line || '')
  logText.value = `${logText.value}${text}\n`
}

function appendOutput(output: string) {
  const text = String(output || '').replace(/\r\n/g, '\n')
  if (!text) return
  const lines = text.endsWith('\n') ? text.slice(0, -1).split('\n') : text.split('\n')
  lines.forEach((line) => appendLogLine(line))
}

function formatNow(): string {
  return new Date().toLocaleString()
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.fields?.command || err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function appendToStash() {
  const lines = parseLines(stashInput.value)
  if (lines.length === 0) {
    ElMessage.warning(t('shell.messages.needAtLeastOne'))
    return
  }

  const invalid = lines.find((line) => !isOpenclawCommand(line))
  if (invalid) {
    ElMessage.error(t('shell.messages.invalidStashCommand', { command: invalid }))
    return
  }

  const existing = new Set(stashCommands.value.map((item) => item.command))
  let added = 0
  lines.forEach((line) => {
    if (!existing.has(line)) {
      stashCommands.value.push({ id: newRowID(), command: line })
      existing.add(line)
      added += 1
    }
  })

  stashInput.value = ''
  if (added === 0) {
    ElMessage.info(t('shell.messages.alreadyInStash'))
    return
  }
  ElMessage.success(t('shell.messages.addedToStash', { count: added }))
}

function removeStash(id: string) {
  stashCommands.value = stashCommands.value.filter((item) => item.id !== id)
}

function clearStash() {
  stashCommands.value = []
}

function fillCommand(command: string) {
  commandInput.value = String(command || '')
}

function clearLog() {
  logText.value = ''
}

async function executeCommand() {
  const command = commandInput.value.trim()
  if (!command) {
    ElMessage.warning(t('shell.messages.needCommand'))
    return
  }
  if (!isOpenclawCommand(command)) {
    ElMessage.error(t('shell.messages.invalidExecuteCommand'))
    return
  }
  if (!canExecute.value) {
    ElMessage.warning(t('shell.messages.noPermission'))
    return
  }

  executing.value = true
  appendLogLine(`[${formatNow()}] $ ${command}`)
  try {
    const result = await executeShellCommand(command)
    appendOutput(result.output)
    if (result.error) {
      appendLogLine(`[${formatNow()}] error: ${result.error}`)
    }
    appendLogLine(`[${formatNow()}] exit=${result.exit_code} success=${result.success} duration=${result.duration_ms}ms`)
  } catch (err) {
    appendLogLine(`[${formatNow()}] ${t('shell.messages.executeFailed')}: ${parseError(err, t('shell.messages.executeFailed'))}`)
  } finally {
    executing.value = false
  }
}

function loadStash() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return
    stashCommands.value = parsed
      .map((item: any) => ({
        id: String(item?.id || newRowID()),
        command: String(item?.command || '').trim(),
      }))
      .filter((item) => item.command && isOpenclawCommand(item.command))
  } catch {
    // ignore malformed cache
  }
}

watch(stashCommands, (value) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
  } catch {
    // ignore storage failure
  }
}, { deep: true })

watch(logText, () => {
  if (!logBoxRef.value) return
  logBoxRef.value.scrollTop = logBoxRef.value.scrollHeight
})

onMounted(loadStash)
</script>

<style scoped>
.shell-page {
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

.panel-title {
  font-weight: 600;
}

.stash-toolbar {
  margin: 10px 0;
}

.security-tip {
  margin-bottom: 10px;
}

.command-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 8px;
  align-items: center;
}

.log-box {
  margin-top: 10px;
  height: 340px;
  overflow: auto;
  background: #101318;
  color: #e5e7eb;
  border-radius: 8px;
  padding: 10px;
  line-height: 1.45;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
}

@media (max-width: 900px) {
  .command-row {
    grid-template-columns: 1fr;
  }
}
</style>
