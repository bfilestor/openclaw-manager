<template>
  <div class="binding-graph-page">
    <div class="topbar">
      <el-space>
        <el-button @click="goBackToAgents">{{ t('bindings.backToAgents') }}</el-button>
        <h3>{{ t('bindings.title') }}</h3>
      </el-space>
      <el-space>
        <el-tag type="info">{{ t('bindings.agentCount', { count: graph.agents.length }) }}</el-tag>
        <el-tag type="info">{{ t('bindings.botCount', { count: graph.bots.length }) }}</el-tag>
        <el-tag type="success">{{ t('bindings.edgeCount', { count: relationEdges.length }) }}</el-tag>
        <el-tag v-if="isEditMode" type="warning">{{ t('bindings.editing') }}</el-tag>
        <el-button :loading="loading" @click="loadGraph">{{ t('common.actions.refresh') }}</el-button>
        <el-button v-if="!isEditMode" type="primary" @click="startEdit">{{ t('bindings.editMode') }}</el-button>
        <template v-else>
          <el-button :disabled="!canUndo" @click="undoLast">{{ t('bindings.undo') }}</el-button>
          <el-button @click="cancelEdit">{{ t('bindings.cancelEdit') }}</el-button>
          <el-button type="primary" :loading="saving" @click="saveBindings">{{ t('bindings.saveBindings') }}</el-button>
        </template>
      </el-space>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ t('bindings.visualTitle') }}</span>
          <el-text type="info">{{ t('bindings.updatedAt', { time: formatDateTime(modifiedAt) }) }}</el-text>
        </div>
      </template>

      <el-empty
        v-if="!loading && relationEdges.length === 0"
        :description="t('bindings.empty')"
      />

      <el-scrollbar v-else class="graph-scroll">
        <div ref="canvasRef" class="graph-canvas" :style="{ height: `${canvasHeight}px` }">
          <div class="column-label left">{{ t('bindings.columns.agents') }}</div>
          <div class="column-label right">{{ t('bindings.columns.bots') }}</div>

          <svg class="edge-layer" :viewBox="`0 0 ${canvasWidth} ${canvasHeight}`" preserveAspectRatio="none">
            <defs>
              <marker
                id="edge-arrow"
                viewBox="0 0 10 10"
                refX="8"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" class="edge-arrow" />
              </marker>
            </defs>

            <g v-for="edge in drawableEdges" :key="edge.id">
              <path
                :d="edge.path"
                class="edge"
                marker-end="url(#edge-arrow)"
              />
              <circle
                v-if="isEditMode"
                class="edge-handle"
                :cx="edge.endX"
                :cy="edge.endY"
                r="7"
                @pointerdown.stop.prevent="startDragExisting(edge, $event)"
              />
              <text
                class="edge-label"
                :x="(edge.startX + edge.endX) / 2"
                :y="(edge.startY + edge.endY) / 2 - 6"
              >
                {{ edge.channel && edge.account ? `${edge.channel}/${edge.account}` : edge.peer || t('bindings.bindingFallback') }}
              </text>
            </g>

            <path
              v-if="draftLine"
              :d="draftLine.path"
              class="edge edge-draft"
              marker-end="url(#edge-arrow)"
            />
          </svg>

          <div
            v-for="agent in graph.agents"
            :key="`agent-${agent.id}`"
            class="node node-agent"
            :style="nodeStyle('agent', agent.id)"
          >
            <div class="node-title">{{ agent.label }}</div>
            <div class="node-subtitle">{{ t('bindings.columns.agent') }}</div>
            <button
              v-if="isEditMode"
              class="agent-drag-btn"
              @pointerdown.stop.prevent="startDragNew(agent.id, $event)"
            >+
            </button>
          </div>

          <div
            v-for="bot in graph.bots"
            :key="`bot-${bot.id}`"
            class="node node-bot"
            :class="{ 'bot-active': activeBotID === bot.id }"
            :style="nodeStyle('bot', bot.id)"
          >
            <div class="node-title">{{ bot.label }}</div>
            <div class="node-subtitle">{{ t('bindings.columns.bot') }}</div>
          </div>
        </div>
      </el-scrollbar>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="detail-header">
          <span>{{ t('bindings.detailsTitle') }}</span>
          <el-space>
            <el-text type="info">{{ t('bindings.filterByChannel') }}</el-text>
            <el-select v-model="channelFilter" style="width: 180px">
              <el-option :label="t('bindings.all')" value="ALL" />
              <el-option
                v-for="channel in channelOptions"
                :key="channel"
                :label="channel"
                :value="channel"
              />
            </el-select>
          </el-space>
        </div>
      </template>
      <el-table :data="filteredEdges" row-key="id" style="width: 100%">
        <el-table-column prop="agent_id" :label="t('bindings.columns.agent')" min-width="180" />
        <el-table-column prop="bot_id" :label="t('bindings.columns.bot')" min-width="220" />
        <el-table-column prop="channel" :label="t('bindings.columns.channel')" min-width="120" />
        <el-table-column prop="account" :label="t('bindings.columns.account')" min-width="120" />
        <el-table-column prop="peer" :label="t('bindings.columns.peer')" min-width="180" />
        <el-table-column prop="source" :label="t('bindings.sourcePath')" min-width="220" />
        <el-table-column v-if="isEditMode" :label="t('bindings.columns.actions')" width="100">
          <template #default="{ row }">
            <el-button type="danger" link @click="removeEdge(row.id)">{{ t('common.actions.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="previewVisible" :title="t('bindings.previewTitle')" width="760px">
      <el-alert :title="t('bindings.previewTip')" type="info" show-icon :closable="false" />
      <el-scrollbar height="320px">
        <pre class="preview-box">{{ previewText }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-space>
          <el-button @click="previewVisible = false">{{ t('common.actions.cancel') }}</el-button>
          <el-button type="primary" :loading="saving" @click="confirmSaveBindings">{{ t('bindings.confirmSave') }}</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

type AnyRecord = Record<string, any>

type GraphNode = {
  id: string
  label: string
}

type GraphEdge = {
  id: string
  agent_id: string
  bot_id: string
  channel: string
  account: string
  peer: string
  source: string
}

type GraphSnapshot = {
  agents: GraphNode[]
  bots: GraphNode[]
  edges: GraphEdge[]
}

type DrawableEdge = GraphEdge & {
  path: string
  startX: number
  startY: number
  endX: number
  endY: number
}

const loading = ref(false)
const errorMessage = ref('')
const modifiedAt = ref('')
const graph = ref<GraphSnapshot>({ agents: [], bots: [], edges: [] })
const editableEdges = ref<GraphEdge[]>([])
const isEditMode = ref(false)
const saving = ref(false)
const rawConfig = ref<any>(null)
const previewVisible = ref(false)
const previewText = ref('')
const channelFilter = ref('ALL')
const router = useRouter()
const { t } = useI18n()

type DraftLine = { path: string }
const draftLine = ref<DraftLine | null>(null)
const activeBotID = ref('')
const dragState = ref<null | { edgeID?: string; agentID: string; pointerId: number }>(null)
const canvasRef = ref<HTMLElement | null>(null)
const historyStack = ref<GraphEdge[][]>([])
const canUndo = computed(() => historyStack.value.length > 0)

const canvasWidth = 1180
const nodeWidth = 250
const nodeHeight = 72
const nodeGap = 92
const topPadding = 52
const bottomPadding = 36
const leftX = 80
const rightX = canvasWidth - nodeWidth - 80

const activeEdges = computed(() => (isEditMode.value ? editableEdges.value : graph.value.edges))

const rows = computed(() => Math.max(graph.value.agents.length, graph.value.bots.length, 1))
const canvasHeight = computed(() => topPadding + bottomPadding + nodeHeight + (rows.value - 1) * nodeGap)

const agentPositionMap = computed(() => {
  const map = new Map<string, { x: number; y: number }>()
  graph.value.agents.forEach((node, index) => {
    map.set(node.id, { x: leftX, y: topPadding + index * nodeGap })
  })
  return map
})

const botPositionMap = computed(() => {
  const map = new Map<string, { x: number; y: number }>()
  graph.value.bots.forEach((node, index) => {
    map.set(node.id, { x: rightX, y: topPadding + index * nodeGap })
  })
  return map
})

const drawableEdges = computed<DrawableEdge[]>(() => {
  const linkEdges: GraphEdge[] = []
  const linkSeen = new Set<string>()
  for (const edge of activeEdges.value) {
    const key = `${edge.agent_id}|${edge.channel}|${edge.account}`
    if (linkSeen.has(key)) continue
    linkSeen.add(key)
    linkEdges.push(edge)
  }

  const totals = new Map<string, number>()
  linkEdges.forEach((edge) => {
    const key = `${edge.agent_id}->${edge.bot_id}`
    totals.set(key, (totals.get(key) || 0) + 1)
  })

  const used = new Map<string, number>()
  const out: DrawableEdge[] = []

  for (const edge of linkEdges) {
    const from = agentPositionMap.value.get(edge.agent_id)
    const to = botPositionMap.value.get(edge.bot_id)
    if (!from || !to) continue

    const pairKey = `${edge.agent_id}->${edge.bot_id}`
    const total = totals.get(pairKey) || 1
    const ordinal = used.get(pairKey) || 0
    used.set(pairKey, ordinal + 1)
    const offset = (ordinal - (total - 1) / 2) * 8

    const startX = from.x + nodeWidth
    const endX = to.x
    const startY = from.y + nodeHeight / 2 + offset
    const endY = to.y + nodeHeight / 2 + offset
    const controlDistance = (endX - startX) * 0.45
    const path = `M ${startX} ${startY} C ${startX + controlDistance} ${startY}, ${endX - controlDistance} ${endY}, ${endX} ${endY}`

    out.push({
      ...edge,
      path,
      startX,
      startY,
      endX,
      endY
    })
  }

  return out
})

const channelOptions = computed(() => {
  const set = new Set<string>()
  for (const edge of activeEdges.value) {
    const channel = String(edge.channel || '').trim()
    if (channel) set.add(channel)
  }
  return Array.from(set).sort((a, b) => a.localeCompare(b))
})

const filteredEdges = computed(() => {
  const selected = channelFilter.value
  if (!selected || selected === 'ALL') return activeEdges.value
  return activeEdges.value.filter((edge) => edge.channel === selected)
})

const relationEdges = computed(() => {
  const seen = new Set<string>()
  const out: GraphEdge[] = []
  for (const edge of activeEdges.value) {
    const key = `${edge.agent_id}|${edge.channel}|${edge.account}`
    if (seen.has(key)) continue
    seen.add(key)
    out.push(edge)
  }
  return out
})

function isRecord(v: unknown): v is AnyRecord {
  return !!v && typeof v === 'object' && !Array.isArray(v)
}

function pickString(obj: AnyRecord, keys: string[]): string {
  for (const key of keys) {
    const raw = obj?.[key]
    if (raw === undefined || raw === null) continue
    const text = String(raw).trim()
    if (text) return text
  }
  return ''
}

function parsePeerText(v: any): string {
  if (v === undefined || v === null) return ''
  if (typeof v === 'string' || typeof v === 'number') return String(v)
  if (isRecord(v)) {
    const kind = pickString(v, ['kind', 'type'])
    const id = pickString(v, ['id', 'peer', 'value'])
    if (kind && id) return `${kind}:${id}`
    if (id) return id
  }
  return ''
}

function extractGraphFromConfig(config: any): GraphSnapshot {
  const agentMap = new Map<string, GraphNode>()
  const botMap = new Map<string, GraphNode>()
  const edges: GraphEdge[] = []
  const edgeSeen = new Set<string>()
  const accountMetaKeys = new Set([
    'enabled',
    'token',
    'proxy',
    'timeout',
    'retries',
    'retry',
    'webhook',
    'api',
    'api_url',
    'base_url',
    'secret',
    'key',
    'type',
    'format',
    'default',
    'default_account',
    'accounts',
    'bots',
    'channel',
    'channels',
    'bindings',
    'rules',
    'map',
    'items',
    'agent',
    'agent_id',
    'agentid',
    'accountid',
    'account_id'
  ])
  const wrapperKeys = ['channels', 'channel', 'bindings', 'rules', 'map', 'items']

  const addAgent = (id: string, label?: string) => {
    const clean = String(id || '').trim()
    if (!clean) return
    if (!agentMap.has(clean)) {
      agentMap.set(clean, { id: clean, label: String(label || clean) })
    }
  }

  const addBot = (id: string, label?: string) => {
    const clean = String(id || '').trim()
    if (!clean) return
    if (!botMap.has(clean)) {
      botMap.set(clean, { id: clean, label: String(label || clean) })
    }
  }

  const addBotByChannelAccount = (channel: string, account: string) => {
    const cleanChannel = String(channel || '').trim()
    const cleanAccount = String(account || '').trim()
    if (!cleanChannel || !cleanAccount) return
    addBot(`${cleanChannel}/${cleanAccount}`, `${cleanChannel}/${cleanAccount}`)
  }

  const addEdge = (payload: {
    agentID: string
    channel: string
    account: string
    peer?: string
    source: string
  }) => {
    const agentID = String(payload.agentID || '').trim()
    const channel = String(payload.channel || '').trim()
    const account = String(payload.account || '').trim()
    if (!agentID || !channel || !account) return
    const botID = `${channel}/${account}`
    addAgent(agentID)
    addBotByChannelAccount(channel, account)
    const peer = String(payload.peer || '')
    const dedupeKey = [agentID, botID, channel, account, peer].join('|')
    if (edgeSeen.has(dedupeKey)) return
    edgeSeen.add(dedupeKey)
    edges.push({
      id: `edge-${edges.length + 1}`,
      agent_id: agentID,
      bot_id: botID,
      channel,
      account,
      peer,
      source: payload.source
    })
  }

  const parseChannelAccounts = (raw: any, source: string) => {
    if (!isRecord(raw)) return

    const parseAccountContainer = (channel: string, container: any, _containerSource: string) => {
      if (Array.isArray(container)) {
        container.forEach((item) => {
          if (typeof item === 'string' && item.trim()) {
            addBotByChannelAccount(channel, item)
            return
          }
          if (!isRecord(item)) return
          const account = pickString(item, ['account', 'name', 'id', 'bot_id'])
          if (account) addBotByChannelAccount(channel, account)
        })
        return
      }
      if (!isRecord(container)) return

      const nestedAccounts = isRecord(container.accounts) ? container.accounts : undefined
      const nestedBots = isRecord(container.bots) ? container.bots : undefined
      if (nestedAccounts) {
        for (const account of Object.keys(nestedAccounts)) {
          addBotByChannelAccount(channel, account)
        }
      }
      if (nestedBots) {
        for (const account of Object.keys(nestedBots)) {
          addBotByChannelAccount(channel, account)
        }
      }
      if (nestedAccounts || nestedBots) return

      const explicitAccount = pickString(container, ['account', 'default_account'])
      if (explicitAccount) addBotByChannelAccount(channel, explicitAccount)
    }

    for (const [channel, channelConfig] of Object.entries(raw)) {
      parseAccountContainer(channel, channelConfig, `${source}.${channel}`)
    }
  }

  const applyBinding = (channel: string, account: string, rawAgent: any, source: string, peer = '') => {
    const agentID = String(rawAgent ?? '').trim()
    if (!agentID) return
    addEdge({ agentID, channel, account, peer, source })
  }

  const parseBindingTarget = (channel: string, account: string, target: any, source: string) => {
    if (target === null || target === undefined) return
    if (typeof target === 'string' || typeof target === 'number') {
      applyBinding(channel, account, target, source)
      return
    }
    if (Array.isArray(target)) {
      target.forEach((item, index) => parseBindingTarget(channel, account, item, `${source}[${index}]`))
      return
    }
    if (!isRecord(target)) return

    const directAgent = pickString(target, ['agent_id', 'agentId', 'agent'])
    const directChannel = pickString(target, ['channel']) || channel
    const directAccount = pickString(target, ['account', 'accountId', 'bot', 'bot_id']) || account
    const peer = parsePeerText(target.peer) || pickString(target, ['target', 'to'])

    if (isRecord(target.match)) {
      const matchChannel = pickString(target.match, ['channel']) || directChannel
      const matchAccount = pickString(target.match, ['accountId', 'account', 'bot', 'bot_id']) || directAccount
      const matchPeer = parsePeerText(target.match.peer) || peer
      if (directAgent && matchChannel && matchAccount) {
        applyBinding(matchChannel, matchAccount, directAgent, source, matchPeer)
        return
      }
    }

    if (directAgent && directChannel && directAccount) {
      applyBinding(directChannel, directAccount, directAgent, source, peer)
      return
    }

    if (Array.isArray(target.agents)) {
      target.agents.forEach((agentItem, index) => {
        const agentID = typeof agentItem === 'string' ? agentItem : pickString(agentItem, ['id', 'agent_id', 'agentId', 'agent'])
        applyBinding(directChannel, directAccount, agentID, `${source}.agents[${index}]`, peer)
      })
    }
  }

  const parseBindingsByChannel = (channel: string, raw: any, source: string) => {
    if (raw === null || raw === undefined) return

    if (Array.isArray(raw)) {
      raw.forEach((item, index) => {
        if (!isRecord(item)) return
        const account = pickString(item, ['account', 'accountId', 'bot', 'bot_id'])
        const agentID = pickString(item, ['agent_id', 'agentId', 'agent'])
        if (account && agentID) {
          applyBinding(channel, account, agentID, `${source}[${index}]`, parsePeerText(item.peer) || pickString(item, ['target', 'to']))
          return
        }
        for (const [accountKey, target] of Object.entries(item)) {
          parseBindingTarget(channel, accountKey, target, `${source}[${index}].${accountKey}`)
        }
      })
      return
    }

    if (!isRecord(raw)) return

    const singleAccount = pickString(raw, ['account', 'accountId', 'bot', 'bot_id'])
    const singleAgent = pickString(raw, ['agent_id', 'agentId', 'agent'])
    if (singleAccount && singleAgent) {
      applyBinding(channel, singleAccount, singleAgent, source, parsePeerText(raw.peer) || pickString(raw, ['target', 'to']))
      return
    }

    for (const [accountKey, target] of Object.entries(raw)) {
      const lk = String(accountKey).toLowerCase()
      if (accountMetaKeys.has(lk) && !isRecord(target) && !Array.isArray(target)) continue
      parseBindingTarget(channel, accountKey, target, `${source}.${accountKey}`)
    }
  }

  const parseBindingsRoot = (raw: any, source: string) => {
    if (raw === null || raw === undefined) return

    if (Array.isArray(raw)) {
      raw.forEach((item, index) => {
        if (!isRecord(item)) return
        const match = isRecord(item.match) ? item.match : undefined
        const channel = match ? pickString(match, ['channel']) : pickString(item, ['channel'])
        const account = match
          ? pickString(match, ['accountId', 'account', 'bot', 'bot_id'])
          : pickString(item, ['accountId', 'account', 'bot', 'bot_id'])
        const agentID = pickString(item, ['agent_id', 'agentId', 'agent'])
        if (channel && account && agentID) {
          const peer = match ? parsePeerText(match.peer) : parsePeerText(item.peer)
          applyBinding(channel, account, agentID, `${source}[${index}]`, peer || pickString(item, ['target', 'to']))
          return
        }
        for (const [channelKey, value] of Object.entries(item)) {
          parseBindingsByChannel(channelKey, value, `${source}[${index}].${channelKey}`)
        }
      })
      return
    }

    if (!isRecord(raw)) return

    const directChannel = pickString(raw, ['channel'])
    const directAccount = pickString(raw, ['account', 'accountId', 'bot', 'bot_id'])
    const directAgent = pickString(raw, ['agent_id', 'agentId', 'agent'])
    if (directChannel && directAccount && directAgent) {
      applyBinding(directChannel, directAccount, directAgent, source, parsePeerText(raw.peer) || pickString(raw, ['target', 'to']))
      return
    }

    for (const wrapper of wrapperKeys) {
      const wrapped = raw[wrapper]
      if (isRecord(wrapped)) {
        for (const [channelKey, channelBindings] of Object.entries(wrapped)) {
          parseBindingsByChannel(channelKey, channelBindings, `${source}.${wrapper}.${channelKey}`)
        }
      } else if (Array.isArray(wrapped)) {
        parseBindingsRoot(wrapped, `${source}.${wrapper}`)
      }
    }

    for (const [channelKey, value] of Object.entries(raw)) {
      if (wrapperKeys.includes(channelKey)) continue
      if (['channel', 'account', 'accountId', 'agent', 'agent_id', 'agentId', 'peer', 'target', 'to', 'match'].includes(channelKey)) continue
      parseBindingsByChannel(channelKey, value, `${source}.${channelKey}`)
    }
  }

  const parseAgents = (raw: any, source: string) => {
    if (raw === null || raw === undefined) return
    if (Array.isArray(raw)) {
      raw.forEach((item, index) => {
        if (!isRecord(item)) return
        const agentID = pickString(item, ['id', 'agent_id', 'agent', 'name'])
        if (!agentID) return
        addAgent(agentID)
        const bindings = item.bindings
        if (!Array.isArray(bindings)) return
        bindings.forEach((binding, bIndex) => {
          if (!isRecord(binding)) return
          const channel = pickString(binding, ['channel'])
          const account = pickString(binding, ['account', 'bot', 'bot_id'])
          if (!channel || !account) return
          addEdge({
            agentID,
            channel,
            account,
            peer: pickString(binding, ['peer', 'target', 'to']),
            source: `${source}[${index}].bindings[${bIndex}]`
          })
        })
      })
      return
    }
    if (!isRecord(raw)) return

    if (Array.isArray(raw.list)) {
      parseAgents(raw.list, `${source}.list`)
      return
    }

    if (Array.isArray(raw.items)) {
      parseAgents(raw.items, `${source}.items`)
      return
    }

    for (const [key, value] of Object.entries(raw)) {
      if (!isRecord(value)) {
        continue
      }

      if (key === 'defaults' || key === 'default') continue

      const agentID = pickString(value, ['id', 'agent_id', 'agent', 'name']) || String(key)
      addAgent(agentID)
      const bindings = value.bindings
      if (!Array.isArray(bindings)) continue
      bindings.forEach((binding, bIndex) => {
        if (!isRecord(binding)) return
        const channel = pickString(binding, ['channel'])
        const account = pickString(binding, ['account', 'bot', 'bot_id'])
        if (!channel || !account) return
        addEdge({
          agentID,
          channel,
          account,
          peer: pickString(binding, ['peer', 'target', 'to']),
          source: `${source}.${agentID}.bindings[${bIndex}]`
        })
      })
    }
  }

  parseChannelAccounts(config?.channels, 'channels')
  parseChannelAccounts(config?.channel, 'channel')
  parseChannelAccounts(config?.bots, 'bots')
  parseChannelAccounts(config?.bot, 'bot')
  parseBindingsRoot(config?.bindings, 'bindings')
  parseAgents(config?.agents, 'agents')

  const agents = Array.from(agentMap.values()).sort((a, b) => a.label.localeCompare(b.label))
  const bots = Array.from(botMap.values()).sort((a, b) => a.label.localeCompare(b.label))
  edges.sort((a, b) => {
    const left = `${a.channel}|${a.account}|${a.agent_id}|${a.peer}`
    const right = `${b.channel}|${b.account}|${b.agent_id}|${b.peer}`
    return left.localeCompare(right)
  })

  return { agents, bots, edges }
}

function nodeStyle(kind: 'agent' | 'bot', id: string): Record<string, string> {
  const hit = kind === 'agent' ? agentPositionMap.value.get(id) : botPositionMap.value.get(id)
  if (!hit) return {}
  return {
    left: `${hit.x}px`,
    top: `${hit.y}px`,
    width: `${nodeWidth}px`,
    height: `${nodeHeight}px`
  }
}

function toCanvasXY(clientX: number, clientY: number) {
  const rect = canvasRef.value?.getBoundingClientRect()
  if (!rect) return { x: clientX, y: clientY }
  return { x: clientX - rect.left, y: clientY - rect.top }
}

function getBotIDFromPoint(clientX: number, clientY: number): string {
  const pt = toCanvasXY(clientX, clientY)
  for (const bot of graph.value.bots) {
    const hit = botPositionMap.value.get(bot.id)
    if (!hit) continue
    const left = hit.x
    const top = hit.y
    const right = left + nodeWidth
    const bottom = top + nodeHeight
    if (pt.x >= left && pt.x <= right && pt.y >= top && pt.y <= bottom) {
      return bot.id
    }
  }
  return ''
}

function updateDraftLine(agentID: string, pointerX: number, pointerY: number) {
  const from = agentPositionMap.value.get(agentID)
  if (!from) return
  const pt = toCanvasXY(pointerX, pointerY)
  const startX = from.x + nodeWidth
  const startY = from.y + nodeHeight / 2
  const endX = pt.x
  const endY = pt.y
  const controlDistance = (endX - startX) * 0.45
  draftLine.value = {
    path: `M ${startX} ${startY} C ${startX + controlDistance} ${startY}, ${endX - controlDistance} ${endY}, ${endX} ${endY}`
  }
}

function pushHistory() {
  historyStack.value.push(editableEdges.value.map((edge) => ({ ...edge })))
  if (historyStack.value.length > 100) historyStack.value.shift()
}

function undoLast() {
  const last = historyStack.value.pop()
  if (!last) return
  editableEdges.value = last.map((edge) => ({ ...edge }))
}

function replaceEdgeTarget(edge: GraphEdge, targetBotID: string) {
  const [channel, account] = targetBotID.split('/')
  if (!channel || !account) return
  edge.bot_id = targetBotID
  edge.channel = channel
  edge.account = account
}

function addEdgeByAgentToBot(agentID: string, targetBotID: string) {
  const [channel, account] = targetBotID.split('/')
  if (!channel || !account) return
  const exists = editableEdges.value.some((edge) => edge.agent_id === agentID && edge.channel === channel && edge.account === account)
  if (exists) return
  editableEdges.value.push({
    id: `edge-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    agent_id: agentID,
    bot_id: targetBotID,
    channel,
    account,
    peer: '',
    source: 'bindings.editor'
  })
}

function finishDrag(targetBotID: string) {
  const state = dragState.value
  if (!state || !targetBotID) return
  pushHistory()
  if (state.edgeID) {
    const edge = editableEdges.value.find((item) => item.id === state.edgeID)
    if (edge) replaceEdgeTarget(edge, targetBotID)
  } else {
    addEdgeByAgentToBot(state.agentID, targetBotID)
  }
}

function onGlobalPointerMove(ev: PointerEvent) {
  if (!dragState.value) return
  updateDraftLine(dragState.value.agentID, ev.clientX, ev.clientY)
  activeBotID.value = getBotIDFromPoint(ev.clientX, ev.clientY)
}

function cleanupDrag() {
  dragState.value = null
  draftLine.value = null
  activeBotID.value = ''
  window.removeEventListener('pointermove', onGlobalPointerMove)
  window.removeEventListener('pointerup', onGlobalPointerUp)
}

function onGlobalPointerUp(ev: PointerEvent) {
  if (!dragState.value) return
  const targetBotID = getBotIDFromPoint(ev.clientX, ev.clientY)
  finishDrag(targetBotID)
  cleanupDrag()
}

function startDragExisting(edge: GraphEdge, ev: PointerEvent) {
  if (!isEditMode.value) return
  dragState.value = { edgeID: edge.id, agentID: edge.agent_id, pointerId: ev.pointerId }
  updateDraftLine(edge.agent_id, ev.clientX, ev.clientY)
  window.addEventListener('pointermove', onGlobalPointerMove)
  window.addEventListener('pointerup', onGlobalPointerUp)
}

function startDragNew(agentID: string, ev: PointerEvent) {
  if (!isEditMode.value) return
  dragState.value = { agentID, pointerId: ev.pointerId }
  updateDraftLine(agentID, ev.clientX, ev.clientY)
  window.addEventListener('pointermove', onGlobalPointerMove)
  window.addEventListener('pointerup', onGlobalPointerUp)
}

function startEdit() {
  isEditMode.value = true
  historyStack.value = []
  editableEdges.value = graph.value.edges.map((edge) => ({ ...edge }))
}

function cancelEdit() {
  cleanupDrag()
  previewVisible.value = false
  isEditMode.value = false
  historyStack.value = []
  editableEdges.value = []
}

function buildBindingsPayload(edges: GraphEdge[]) {
  const bindings: Record<string, Record<string, any>> = {}
  for (const edge of edges) {
    if (!bindings[edge.channel]) bindings[edge.channel] = {}
    const current = bindings[edge.channel][edge.account]
    const value = edge.peer ? { agent_id: edge.agent_id, peer: edge.peer } : edge.agent_id
    if (!current) {
      bindings[edge.channel][edge.account] = value
      continue
    }
    if (Array.isArray(current)) {
      current.push(value)
      continue
    }
    bindings[edge.channel][edge.account] = [current, value]
  }
  return bindings
}

function removeEdge(edgeID: string) {
  pushHistory()
  editableEdges.value = editableEdges.value.filter((edge) => edge.id !== edgeID)
}

function buildPreviewText() {
  const oldBindings = rawConfig.value?.bindings ?? {}
  const newBindings = buildBindingsPayload(editableEdges.value)
  const oldText = JSON.stringify(oldBindings, null, 2)
  const newText = JSON.stringify(newBindings, null, 2)
  previewText.value = `--- old bindings\n${oldText}\n\n--- new bindings\n${newText}`
}

function saveBindings() {
  if (!rawConfig.value) {
    ElMessage.error(t('bindings.messages.noConfigLoaded'))
    return
  }
  buildPreviewText()
  previewVisible.value = true
}

async function confirmSaveBindings() {
  if (!rawConfig.value) {
    ElMessage.error(t('bindings.messages.noConfigLoaded'))
    return
  }
  saving.value = true
  try {
    const next = JSON.parse(JSON.stringify(rawConfig.value))
    next.bindings = buildBindingsPayload(editableEdges.value)
    const formatted = JSON.stringify(next, null, 2)
    await axios.put('/api/v1/config/openclaw', { content: formatted })
    ElMessage.success(t('bindings.messages.saveSuccess'))
    previewVisible.value = false
    isEditMode.value = false
    historyStack.value = []
    await loadGraph()
  } catch (err) {
    ElMessage.error(parseError(err, t('bindings.messages.saveFailed')))
  } finally {
    saving.value = false
  }
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function formatDateTime(value: string): string {
  if (!value) return t('common.emptyValue')
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString()
}

async function loadGraph() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get('/api/v1/config/openclaw')
    modifiedAt.value = String(data?.modified_at || '')
    const rawContent = String(data?.content || '').trim()
    if (!rawContent) {
      graph.value = { agents: [], bots: [], edges: [] }
      return
    }
    let parsed: any
    try {
      parsed = JSON.parse(rawContent)
    } catch {
      throw new Error(t('bindings.messages.parseConfigFailed'))
    }
    rawConfig.value = parsed
    graph.value = extractGraphFromConfig(parsed)
    if (isEditMode.value) {
      editableEdges.value = graph.value.edges.map((edge) => ({ ...edge }))
    }
    if (channelFilter.value !== 'ALL' && !graph.value.edges.some((edge) => edge.channel === channelFilter.value)) {
      channelFilter.value = 'ALL'
    }
  } catch (err) {
    rawConfig.value = null
    graph.value = { agents: [], bots: [], edges: [] }
    channelFilter.value = 'ALL'
    errorMessage.value = parseError(err, t('bindings.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

function goBackToAgents() {
  router.push('/agents')
}

onMounted(loadGraph)
onBeforeUnmount(() => {
  cleanupDrag()
})
</script>

<style scoped>
.binding-graph-page {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.graph-scroll {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  background:
    radial-gradient(1200px 360px at 12% -10%, rgba(30, 136, 229, 0.12), transparent 60%),
    radial-gradient(1200px 360px at 88% 110%, rgba(255, 152, 0, 0.12), transparent 58%),
    #fff;
}

.graph-canvas {
  position: relative;
  width: 1180px;
  min-height: 360px;
}

.edge-layer {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: visible;
  pointer-events: auto;
}

.edge {
  fill: none;
  stroke: #2c6fd8;
  stroke-opacity: 0.58;
  stroke-width: 2.1;
  pointer-events: none;
}

.edge-draft {
  stroke-dasharray: 6 4;
  stroke-opacity: 0.9;
}

.edge-handle {
  fill: #fff;
  stroke: #2c6fd8;
  stroke-width: 2;
  cursor: grab;
  pointer-events: auto;
}

.edge-arrow {
  fill: #2c6fd8;
  opacity: 0.68;
}

.edge-label {
  fill: #5c6470;
  font-size: 11px;
  letter-spacing: 0.2px;
}

.column-label {
  position: absolute;
  top: 16px;
  font-size: 12px;
  color: #606a77;
  letter-spacing: 0.8px;
}

.column-label.left {
  left: 88px;
}

.column-label.right {
  right: 88px;
}

.node {
  position: absolute;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(17, 24, 39, 0.08);
  border: 1px solid #dce3ef;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 10px 12px;
  overflow: hidden;
}

.node-title {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.node-subtitle {
  font-size: 11px;
  color: #68707e;
  margin-top: 4px;
}

.node-agent {
  background: linear-gradient(135deg, #eff7ff, #f8fbff);
}

.node-bot {
  background: linear-gradient(135deg, #fff6ec, #fffdf8);
}

.node-bot.bot-active {
  border-color: #2c6fd8;
  box-shadow: 0 0 0 3px rgba(44, 111, 216, 0.2);
}

.agent-drag-btn {
  position: absolute;
  right: 8px;
  bottom: 8px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: none;
  background: #2c6fd8;
  color: #fff;
  font-weight: 700;
  cursor: crosshair;
}

.preview-box {
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #374151;
  padding: 10px;
}
</style>
