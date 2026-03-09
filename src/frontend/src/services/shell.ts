import axios from 'axios'

export type ShellExecuteResponse = {
  command: string
  output: string
  exit_code: number
  success: boolean
  error?: string
  duration_ms: number
  started_at: string
  finished_at: string
}

export async function executeShellCommand(command: string): Promise<ShellExecuteResponse> {
  const { data } = await axios.post('/api/v1/tasks/shell/execute', { command })
  return {
    command: String(data?.command || command),
    output: String(data?.output || ''),
    exit_code: Number(data?.exit_code ?? -1),
    success: !!data?.success,
    error: typeof data?.error === 'string' ? data.error : '',
    duration_ms: Number(data?.duration_ms ?? 0),
    started_at: String(data?.started_at || ''),
    finished_at: String(data?.finished_at || ''),
  }
}
