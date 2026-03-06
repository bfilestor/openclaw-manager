<template>
  <div class="binding-graph-page">
    <div class="topbar">
      <el-space>
        <el-button @click="goBackToAgents">返回 Agents</el-button>
        <h3>Bindings 图谱</h3>
      </el-space>
      <el-space>
        <el-tag type="info">Agent: {{ graph.agents.length }}</el-tag>
        <el-tag type="info">Bot: {{ graph.bots.length }}</el-tag>
        <el-tag type="success">连线: {{ graph.edges.length }}</el-tag>
        <el-button :loading="loading" @click="loadGraph">刷新</el-button>
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
          <span>openclaw.json 可视化连线</span>
          <el-text type="info">更新时间: {{ formatDateTime(modifiedAt) }}</el-text>
        </div>
      </template>

      <el-empty
        v-if="!loading && graph.edges.length === 0"
        description="未在 openclaw 配置中识别到可展示的 bindings 关系"
      />

      <el-scrollbar v-else class="graph-scroll">
        <div class="graph-canvas" :style="{ height: `${canvasHeight}px` }">
          <div class="column-label left">Agents</div>
          <div class="column-label right">Bots</div>

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
              <text
                class="edge-label"
                :x="(edge.startX + edge.endX) / 2"
                :y="(edge.startY + edge.endY) / 2 - 6"
              >
                {{ edge.channel && edge.account ? `${edge.channel}/${edge.account}` : edge.peer || 'binding' }}
              </text>
            </g>
          </svg>

          <div
            v-for="agent in graph.agents"
            :key="`agent-${agent.id}`"
            class="node node-agent"
            :style="nodeStyle('agent', agent.id)"
          >
            <div class="node-title">{{ agent.label }}</div>
            <div class="node-subtitle">Agent</div>
          </div>

          <div
            v-for="bot in graph.bots"
            :key="`bot-${bot.id}`"
            class="node node-bot"
            :style="nodeStyle('bot', bot.id)"
          >
            <div class="node-title">{{ bot.label }}</div>
            <div class="node-subtitle">Bot</div>
          </div>
        </div>
      </el-scrollbar>
    </el-card>

    <el-card shadow="never">
      <template #header>Binding 明细</template>
      <el-table :data="graph.edges" row-key="id" style="width: 100%">
        <el-table-column prop="agent_id" label="Agent" min-width="180" />
        <el-table-column prop="bot_id" label="Bot" min-width="220" />
        <el-table-column prop="channel" label="Channel" min-width="120" />
        <el-table-column prop="account" label="Account" min-width="120" />
        <el-table-column prop="peer" label="Peer" min-width="180" />
        <el-table-column prop="source" label="来源路径" min-width="220" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

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
const router = useRouter()

const canvasWidth = 1180
const nodeWidth = 250
const nodeHeight = 72
const nodeGap = 92
const topPadding = 52
const bottomPadding = 36
const leftX = 80
const rightX = canvasWidth - nodeWidth - 80

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
  const totals = new Map<string, number>()
  graph.value.edges.forEach((edge) => {
    const key = `${edge.agent_id}->${edge.bot_id}`
    totals.set(key, (totals.get(key) || 0) + 1)
  })

  const used = new Map<string, number>()
  const out: DrawableEdge[] = []

  for (const edge of graph.value.edges) {
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

function isBindingObject(v: AnyRecord): boolean {
  const hasBotRef = Boolean(pickString(v, ['bot_id', 'bot', 'peer', 'target', 'to']))
  const hasChannel = Boolean(pickString(v, ['channel']))
  const hasAccount = Boolean(pickString(v, ['account']))
  return hasBotRef || (hasChannel && hasAccount)
}

function buildBotRef(binding: AnyRecord): { id: string; channel: string; account: string; peer: string } | null {
  const channel = pickString(binding, ['channel'])
  const account = pickString(binding, ['account'])
  const botID = pickString(binding, ['bot_id', 'bot'])
  const peer = pickString(binding, ['peer', 'target', 'to'])

  if (botID) return { id: botID, channel, account, peer }
  if (channel && account) return { id: `${channel}/${account}`, channel, account, peer }
  if (channel && peer) return { id: `${channel}/${peer}`, channel, account, peer }
  if (peer) return { id: peer, channel, account, peer }
  return null
}

function extractGraphFromConfig(config: any): GraphSnapshot {
  const agentMap = new Map<string, GraphNode>()
  const botMap = new Map<string, GraphNode>()
  const edges: GraphEdge[] = []
  const edgeSeen = new Set<string>()

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

  const addEdge = (payload: {
    agentID: string
    botID: string
    channel?: string
    account?: string
    peer?: string
    source: string
  }) => {
    const agentID = payload.agentID.trim()
    const botID = payload.botID.trim()
    if (!agentID || !botID) return
    addAgent(agentID)
    addBot(botID)
    const channel = String(payload.channel || '')
    const account = String(payload.account || '')
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

  const parseBindings = (raw: any, inheritedAgentID: string, source: string) => {
    if (raw === null || raw === undefined) return
    if (Array.isArray(raw)) {
      raw.forEach((item, index) => parseBindings(item, inheritedAgentID, `${source}[${index}]`))
      return
    }
    if (typeof raw === 'string') {
      if (!inheritedAgentID) return
      addEdge({ agentID: inheritedAgentID, botID: raw, source })
      return
    }
    if (typeof raw === 'number' || typeof raw === 'boolean') return
    if (!isRecord(raw)) return

    if (isBindingObject(raw)) {
      const agentID = pickString(raw, ['agent_id', 'agent']) || inheritedAgentID
      const bot = buildBotRef(raw)
      if (agentID && bot) {
        addEdge({
          agentID,
          botID: bot.id,
          channel: bot.channel,
          account: bot.account,
          peer: bot.peer,
          source
        })
      }
      return
    }

    for (const [key, value] of Object.entries(raw)) {
      const nextAgentID = inheritedAgentID || String(key)
      parseBindings(value, nextAgentID, `${source}.${key}`)
    }
  }

  const parseAgents = (raw: any) => {
    if (raw === null || raw === undefined) return
    if (Array.isArray(raw)) {
      raw.forEach((item) => {
        if (!isRecord(item)) return
        const agentID = pickString(item, ['id', 'agent_id', 'agent', 'name'])
        if (!agentID) return
        addAgent(agentID)
        parseBindings(item.bindings, agentID, `agents.${agentID}.bindings`)
      })
      return
    }
    if (!isRecord(raw)) return
    for (const [key, value] of Object.entries(raw)) {
      if (!isRecord(value)) {
        addAgent(String(key))
        continue
      }
      const agentID = pickString(value, ['id', 'agent_id', 'agent', 'name']) || String(key)
      addAgent(agentID)
      parseBindings(value.bindings, agentID, `agents.${agentID}.bindings`)
    }
  }

  const parseBots = (raw: any) => {
    if (raw === null || raw === undefined) return
    if (Array.isArray(raw)) {
      raw.forEach((item) => {
        if (typeof item === 'string') {
          addBot(item)
          return
        }
        if (!isRecord(item)) return
        const channel = pickString(item, ['channel'])
        const account = pickString(item, ['account'])
        const explicitID = pickString(item, ['id', 'bot_id', 'name'])
        if (channel && account) {
          addBot(`${channel}/${account}`)
          return
        }
        if (explicitID) {
          addBot(explicitID)
        }
      })
      return
    }
    if (!isRecord(raw)) return

    for (const [key, value] of Object.entries(raw)) {
      if (!isRecord(value)) {
        addBot(String(key))
        continue
      }

      const explicitID = pickString(value, ['id', 'bot_id', 'name'])
      const channel = pickString(value, ['channel'])
      const account = pickString(value, ['account'])
      if (channel && account) {
        addBot(`${channel}/${account}`)
        continue
      }
      if (explicitID) {
        addBot(explicitID)
        continue
      }

      // 扁平结构: bots.<bot_id> = { channel: "...", token: "..." }
      if (channel) {
        addBot(`${channel}/${key}`)
        continue
      }

      // 嵌套结构: bots.<channel>.<account> = {...}
      const accountKeys = Object.keys(value)
      if (accountKeys.length === 0) {
        addBot(String(key))
        continue
      }
      for (const accountKey of accountKeys) {
        const child = value[accountKey]
        if (isRecord(child)) {
          const childChannel = pickString(child, ['channel']) || key
          const childAccount = pickString(child, ['account']) || accountKey
          addBot(`${childChannel}/${childAccount}`)
          continue
        }
        if (typeof child === 'string' && child.trim()) {
          addBot(`${key}/${accountKey}`)
        }
      }
    }
  }

  parseAgents(config?.agents)
  parseBindings(config?.bindings, '', 'bindings')
  parseBots(config?.bots)

  const agents = Array.from(agentMap.values()).sort((a, b) => a.label.localeCompare(b.label))
  const bots = Array.from(botMap.values()).sort((a, b) => a.label.localeCompare(b.label))
  edges.sort((a, b) => {
    const left = `${a.agent_id}|${a.bot_id}|${a.channel}|${a.account}|${a.peer}`
    const right = `${b.agent_id}|${b.bot_id}|${b.channel}|${b.account}|${b.peer}`
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

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function formatDateTime(value: string): string {
  if (!value) return '-'
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
      throw new Error('openclaw.json 解析失败，请先修复配置 JSON 格式')
    }
    graph.value = extractGraphFromConfig(parsed)
  } catch (err) {
    graph.value = { agents: [], bots: [], edges: [] }
    errorMessage.value = parseError(err, '加载配置失败，无法生成绑定图谱')
  } finally {
    loading.value = false
  }
}

function goBackToAgents() {
  router.push('/agents')
}

onMounted(loadGraph)
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
  pointer-events: none;
}

.edge {
  fill: none;
  stroke: #2c6fd8;
  stroke-opacity: 0.58;
  stroke-width: 2.1;
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
</style>
