import axios from 'axios'

export type BotUsageRow = {
  botId: string
  sessions: number
  inputTokens: number
  outputTokens: number
  totalTokens: number
  estimatedCost: number
}

export type TokenUsageSummary = {
  total: {
    inputTokens: number
    outputTokens: number
    totalTokens: number
    estimatedCost: number
  }
  bots: BotUsageRow[]
  quota?: {
    accountId: string
    tokenLimit: number
    usedTokens: number
    ratio: number
    status: 'normal' | 'near' | 'exceeded'
  }
}

export type ConversationItem = {
  sessionKey: string
  sessionId: string
  agentId: string
  updatedAt: string
  modelProvider: string
  model: string
  inputTokens: number
  outputTokens: number
  totalTokens: number
  estimatedCost: number
  preview: string
}

export type BotConversationPage = {
  botId: string
  total: number
  page: number
  pageSize: number
  items: ConversationItem[]
}

export type SessionMessage = {
  role: string
  timestamp: string
  text: string
}

export async function getTokenUsageSummary(days = 0): Promise<TokenUsageSummary> {
  const { data } = await axios.get('/api/v1/token-usage/summary', {
    params: { days },
  })
  return {
    total: {
      inputTokens: Number(data?.total?.inputTokens || 0),
      outputTokens: Number(data?.total?.outputTokens || 0),
      totalTokens: Number(data?.total?.totalTokens || 0),
      estimatedCost: Number(data?.total?.estimatedCost || 0),
    },
    bots: Array.isArray(data?.bots)
      ? data.bots.map((row: any) => ({
          botId: String(row?.botId || ''),
          sessions: Number(row?.sessions || 0),
          inputTokens: Number(row?.inputTokens || 0),
          outputTokens: Number(row?.outputTokens || 0),
          totalTokens: Number(row?.totalTokens || 0),
          estimatedCost: Number(row?.estimatedCost || 0),
        }))
      : [],
    quota: data?.quota
      ? {
          accountId: String(data?.quota?.accountId || ''),
          tokenLimit: Number(data?.quota?.tokenLimit || 0),
          usedTokens: Number(data?.quota?.usedTokens || 0),
          ratio: Number(data?.quota?.ratio || 0),
          status: (['normal', 'near', 'exceeded'].includes(String(data?.quota?.status))
            ? String(data?.quota?.status)
            : 'normal') as 'normal' | 'near' | 'exceeded',
        }
      : undefined,
  }
}

export async function getBotConversations(botId: string, page = 1, pageSize = 20, days = 0): Promise<BotConversationPage> {
  const { data } = await axios.get(`/api/v1/token-usage/bots/${encodeURIComponent(botId)}/conversations`, {
    params: {
      page,
      page_size: pageSize,
      days,
    },
  })

  return {
    botId: String(data?.botId || botId),
    total: Number(data?.total || 0),
    page: Number(data?.page || page),
    pageSize: Number(data?.pageSize || pageSize),
    items: Array.isArray(data?.items)
      ? data.items.map((row: any) => ({
          sessionKey: String(row?.sessionKey || ''),
          sessionId: String(row?.sessionId || ''),
          agentId: String(row?.agentId || ''),
          updatedAt: String(row?.updatedAt || ''),
          modelProvider: String(row?.modelProvider || ''),
          model: String(row?.model || ''),
          inputTokens: Number(row?.inputTokens || 0),
          outputTokens: Number(row?.outputTokens || 0),
          totalTokens: Number(row?.totalTokens || 0),
          estimatedCost: Number(row?.estimatedCost || 0),
          preview: String(row?.preview || ''),
        }))
      : [],
  }
}

export async function getSessionMessages(sessionId: string, limit = 80): Promise<SessionMessage[]> {
  const { data } = await axios.get(`/api/v1/token-usage/sessions/${encodeURIComponent(sessionId)}/messages`, {
    params: { limit },
  })
  return Array.isArray(data?.items)
    ? data.items.map((row: any) => ({
        role: String(row?.role || 'unknown'),
        timestamp: String(row?.timestamp || ''),
        text: String(row?.text || ''),
      }))
    : []
}
