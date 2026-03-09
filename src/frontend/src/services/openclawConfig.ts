import axios from 'axios'

export type OpenclawConfigPayload = {
  content: string
  size?: number
  modified_at?: string
}

export async function getOpenclawConfig(): Promise<OpenclawConfigPayload> {
  const { data } = await axios.get('/api/v1/config/openclaw')
  return {
    content: typeof data?.content === 'string' ? data.content : '{}',
    size: Number(data?.size || 0),
    modified_at: String(data?.modified_at || ''),
  }
}

export function normalizeOpenclawJSON(raw: string): string {
  const parsed = JSON.parse(raw)
  return JSON.stringify(parsed, null, 2)
}

// revision 保存逻辑统一走后端 PUT /api/v1/config/openclaw
export async function saveOpenclawConfig(content: string): Promise<void> {
  await axios.put('/api/v1/config/openclaw', { content })
}

export function buildOpenclawDiff(fromText: string, toText: string) {
  return {
    fromText: fromText || '{}',
    toText: toText || '{}',
  }
}
