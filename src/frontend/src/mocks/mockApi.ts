import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { gatewayState, usersMock, tasksMock, skillsMock, agentsMock, backupsMock } from './pageMocks'

const mockToken = 'mock.jwt.token'
const mockUser = { user_id: 'u-admin-001', username: 'admin', role: 'Admin' as const }
let mockBackups = [...backupsMock.backups]
let mockTasks = [...tasksMock.tasks]
let mockUsers = [...usersMock.users]
let mockOpenclawConfig = JSON.stringify({
  gateway: {
    bind_addr: '127.0.0.1',
    port: 18790
  },
  channels: {
    telegram: {
      enabled: true,
      botToken: '***',
      groups: {
        '*': { requireMention: true }
      },
      allowFrom: ['1770883652'],
      groupPolicy: 'open',
      streaming: 'partial',
      main_bot: {
        token: '***'
      },
      xcoder_bot: {
        token: '***'
      },
      pmanager_bot: {
        token: '***'
      },
      accounts: {
        main_bot: {
          botToken: '***'
        },
        xcoder_bot: {
          botToken: '***'
        },
        pmanager_bot: {
          botToken: '***'
        }
      }
    }
  },
  agents: {
    defaults: {
      workspace: '/home/openclaw/.openclaw/workspace'
    },
    list: [
      { id: 'main' },
      { id: 'xcoder' },
      { id: 'pmanger' }
    ]
  },
  bindings: [
    {
      agentId: 'main',
      match: {
        channel: 'telegram',
        accountId: 'main_bot',
        peer: { kind: 'direct', id: '1077083652' }
      },
    },
    {
      agentId: 'xcoder',
      match: {
        channel: 'telegram',
        accountId: 'xcoder_bot',
        peer: { kind: 'group', id: '-5757757833' }
      },
    },
    {
      agentId: 'pmanger',
      match: {
        channel: 'telegram',
        accountId: 'pmanager_bot',
        peer: { kind: 'channel', id: '-5757757833' }
      }
    },
  ],
  manager: {
    log_level: 'info'
  }
}, null, 2)
let mockConfigModifiedAt = new Date().toISOString()
let mockConfigRevisions = [
  {
    revision_id: `rev-${Date.now()}-0`,
    target_type: 'openclaw_json',
    target_id: '',
    content: mockOpenclawConfig,
    sha256: 'mock-sha-init',
    created_at: mockConfigModifiedAt,
    created_by: mockUser.user_id
  }
]

function jsonResponse(config: AxiosRequestConfig, data: any, status = 200): Promise<AxiosResponse> {
  return Promise.resolve({
    data,
    status,
    statusText: status >= 400 ? 'ERROR' : 'OK',
    headers: {},
    config
  } as AxiosResponse)
}

function normalizeUrl(url?: string) {
  const raw = String(url || '')
  const q = raw.indexOf('?')
  return q >= 0 ? raw.slice(0, q) : raw
}

function parseRequestData(data: any): any {
  if (!data) return {}
  if (typeof data === 'string') {
    try {
      return JSON.parse(data)
    } catch {
      return {}
    }
  }
  return data
}

function roleWeight(role: string): number {
  if (role === 'Admin') return 3
  if (role === 'Operator') return 2
  return 1
}

function toMockSHA(content: string): string {
  const base = String(content || '')
  const n = base.length
  return `mock-${n.toString(16)}-${Date.now().toString(16)}`
}

export function setupMockApi() {
  axios.interceptors.request.use((config) => {
    const method = String(config.method || 'get').toLowerCase()
    const url = normalizeUrl(config.url)

    const shouldMock =
      url === '/api/v1/auth/login' ||
      url === '/api/v1/auth/refresh' ||
      url === '/api/v1/gateway/status' ||
      url === '/api/v1/gateway/start' ||
      url === '/api/v1/gateway/stop' ||
      url === '/api/v1/gateway/restart' ||
      url === '/api/v1/skills' ||
      url === '/api/v1/agents' ||
      url.startsWith('/api/v1/config/openclaw') ||
      url.startsWith('/api/v1/backups') ||
      url.startsWith('/api/v1/users') ||
      url.startsWith('/api/v1/tasks')

    if (!shouldMock) return config

    config.adapter = async () => {
      // auth
      if (url === '/api/v1/auth/login' && method === 'post') {
        return jsonResponse(config, {
          access_token: mockToken,
          expires_in: 900,
          token_type: 'Bearer',
          user: mockUser
        }) as any
      }
      if (url === '/api/v1/auth/refresh' && method === 'post') {
        return jsonResponse(config, {
          access_token: mockToken,
          expires_in: 900,
          token_type: 'Bearer'
        }) as any
      }

      // dashboard + operation buttons
      if (url === '/api/v1/gateway/status' && method === 'get') {
        return jsonResponse(config, gatewayState) as any
      }
      if (url === '/api/v1/gateway/start' && method === 'post') {
        gatewayState.service.active_state = 'active'
        gatewayState.service.sub_state = 'running'
        gatewayState.service.main_pid = String(Math.floor(10000 + Math.random() * 80000))
        gatewayState.service.active_enter_timestamp = new Date().toISOString()
        return jsonResponse(config, { task_id: 'start-task', status: 'PENDING' }, 202) as any
      }
      if (url === '/api/v1/gateway/stop' && method === 'post') {
        gatewayState.service.active_state = 'inactive'
        gatewayState.service.sub_state = 'dead'
        gatewayState.service.main_pid = '0'
        gatewayState.service.active_enter_timestamp = new Date().toISOString()
        return jsonResponse(config, { task_id: 'stop-task', status: 'PENDING' }, 202) as any
      }
      if (url === '/api/v1/gateway/restart' && method === 'post') {
        gatewayState.service.active_state = 'active'
        gatewayState.service.sub_state = 'running'
        gatewayState.service.main_pid = String(Math.floor(10000 + Math.random() * 80000))
        gatewayState.service.active_enter_timestamp = new Date().toISOString()
        return jsonResponse(config, { task_id: 'restart-task', status: 'PENDING' }, 202) as any
      }

      // users
      if (url === '/api/v1/users' && method === 'get') {
        return jsonResponse(config, { users: mockUsers, total: mockUsers.length }) as any
      }
      if (url === '/api/v1/users/me' && method === 'get') {
        const me = mockUsers.find((u) => u.user_id === mockUser.user_id) || mockUser
        return jsonResponse(config, me) as any
      }
      if (url === '/api/v1/users' && method === 'post') {
        const body = parseRequestData(config.data)
        const username = String(body?.username || '').trim()
        const password = String(body?.password || '')
        const role = String(body?.role || 'Viewer')
        if (!username || !password) {
          return jsonResponse(config, { message: 'username/password required' }, 400) as any
        }
        if (mockUsers.some((u) => u.username === username)) {
          return jsonResponse(config, { message: 'username exists', code: 'USERNAME_EXISTS' }, 409) as any
        }
        const item = {
          user_id: `u-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
          username,
          role: (role === 'Admin' || role === 'Operator' || role === 'Viewer') ? role : 'Viewer',
          status: 'active',
          created_at: new Date().toISOString()
        }
        mockUsers = [...mockUsers, item]
        return jsonResponse(config, item, 201) as any
      }
      const userRoleMatch = url.match(/^\/api\/v1\/users\/([^/]+)\/role$/)
      if (userRoleMatch && method === 'put') {
        const userID = userRoleMatch[1]
        const body = parseRequestData(config.data)
        const role = String(body?.role || '')
        if (!['Viewer', 'Operator', 'Admin'].includes(role)) {
          return jsonResponse(config, { message: 'invalid role' }, 400) as any
        }
        mockUsers = mockUsers.map((u) => u.user_id === userID ? { ...u, role } : u)
        const hit = mockUsers.find((u) => u.user_id === userID)
        if (!hit) return jsonResponse(config, { message: 'user not found' }, 404) as any
        return jsonResponse(config, hit) as any
      }
      const userDisableMatch = url.match(/^\/api\/v1\/users\/([^/]+)\/disable$/)
      if (userDisableMatch && method === 'post') {
        const userID = userDisableMatch[1]
        const body = parseRequestData(config.data)
        const disabled = !!body?.disabled
        mockUsers = mockUsers.map((u) => u.user_id === userID ? { ...u, status: disabled ? 'disabled' : 'active' } : u)
        const hit = mockUsers.find((u) => u.user_id === userID)
        if (!hit) return jsonResponse(config, { message: 'user not found' }, 404) as any
        return jsonResponse(config, hit) as any
      }
      const userPwdMatch = url.match(/^\/api\/v1\/users\/([^/]+)\/password$/)
      if (userPwdMatch && method === 'put') {
        const userID = userPwdMatch[1]
        if (userID === 'me') {
          return jsonResponse(config, { message: 'password updated' }) as any
        }
        const hit = mockUsers.find((u) => u.user_id === userID)
        if (!hit) return jsonResponse(config, { message: 'user not found' }, 404) as any
        return jsonResponse(config, { message: 'password updated' }) as any
      }
      const userDeleteMatch = url.match(/^\/api\/v1\/users\/([^/]+)$/)
      if (userDeleteMatch && method === 'delete') {
        const userID = userDeleteMatch[1]
        const hit = mockUsers.find((u) => u.user_id === userID)
        if (!hit) return jsonResponse(config, { message: 'user not found' }, 404) as any
        if (roleWeight(String(hit.role)) >= roleWeight('Admin')) {
          const admins = mockUsers.filter((u) => u.role === 'Admin').length
          if (admins <= 1) {
            return jsonResponse(config, { message: 'at least two admins required before deleting admin', code: 'LAST_ADMIN_PROTECTED' }, 400) as any
          }
        }
        mockUsers = mockUsers.filter((u) => u.user_id !== userID)
        return jsonResponse(config, { message: 'deleted' }) as any
      }

      // tasks
      if (url === '/api/v1/tasks' && method === 'get') {
        return jsonResponse(config, { tasks: mockTasks, total: mockTasks.length }) as any
      }
      const taskByIDMatch = url.match(/^\/api\/v1\/tasks\/([^/]+)$/)
      if (taskByIDMatch && method === 'get') {
        const item = mockTasks.find((t) => t.task_id === taskByIDMatch[1])
        if (!item) return jsonResponse(config, { error: 'task not found' }, 404) as any
        return jsonResponse(config, item) as any
      }

      // skills
      if (url === '/api/v1/skills' && method === 'get') {
        return jsonResponse(config, skillsMock) as any
      }

      // agents
      if (url === '/api/v1/agents' && method === 'get') {
        return jsonResponse(config, agentsMock) as any
      }

      // config openclaw
      if (url === '/api/v1/config/openclaw' && method === 'get') {
        return jsonResponse(config, {
          content: mockOpenclawConfig,
          size: new Blob([mockOpenclawConfig]).size,
          modified_at: mockConfigModifiedAt
        }) as any
      }
      if (url === '/api/v1/config/openclaw' && method === 'put') {
        const body = parseRequestData(config.data)
        const nextContent = String(body?.content || '')
        try {
          JSON.parse(nextContent)
        } catch {
          return jsonResponse(config, { message: 'invalid json', code: 'INVALID_JSON' }, 400) as any
        }
        mockOpenclawConfig = nextContent
        mockConfigModifiedAt = new Date().toISOString()
        mockConfigRevisions = [
          {
            revision_id: `rev-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
            target_type: 'openclaw_json',
            target_id: '',
            content: mockOpenclawConfig,
            sha256: toMockSHA(mockOpenclawConfig),
            created_at: mockConfigModifiedAt,
            created_by: mockUser.user_id
          },
          ...mockConfigRevisions
        ].slice(0, 50)
        return jsonResponse(config, { message: 'ok' }) as any
      }
      if (url === '/api/v1/config/openclaw/revisions' && method === 'get') {
        return jsonResponse(config, { revisions: mockConfigRevisions }) as any
      }
      const configRestoreMatch = url.match(/^\/api\/v1\/config\/openclaw\/revisions\/([^/]+)\/restore$/)
      if (configRestoreMatch && method === 'post') {
        const revID = configRestoreMatch[1]
        const hit = mockConfigRevisions.find((x) => x.revision_id === revID)
        if (!hit) {
          return jsonResponse(config, { message: 'revision not found' }, 404) as any
        }
        mockOpenclawConfig = String(hit.content || '{}')
        mockConfigModifiedAt = new Date().toISOString()
        mockConfigRevisions = [
          {
            revision_id: `rev-${Date.now()}-${Math.floor(Math.random() * 1000)}`,
            target_type: 'openclaw_json',
            target_id: '',
            content: mockOpenclawConfig,
            sha256: toMockSHA(mockOpenclawConfig),
            created_at: mockConfigModifiedAt,
            created_by: mockUser.user_id
          },
          ...mockConfigRevisions
        ].slice(0, 50)
        return jsonResponse(config, { message: 'restored' }) as any
      }
      const configDeleteMatch = url.match(/^\/api\/v1\/config\/openclaw\/revisions\/([^/]+)$/)
      if (configDeleteMatch && method === 'delete') {
        const revID = configDeleteMatch[1]
        const before = mockConfigRevisions.length
        mockConfigRevisions = mockConfigRevisions.filter((x) => x.revision_id !== revID)
        if (mockConfigRevisions.length === before) {
          return jsonResponse(config, { message: 'revision not found' }, 404) as any
        }
        return jsonResponse(config, { message: 'deleted' }) as any
      }

      // backups
      if (url === '/api/v1/backups' && method === 'get') {
        return jsonResponse(config, { backups: mockBackups }) as any
      }
      if (url === '/api/v1/backups' && method === 'post') {
        const body = parseRequestData(config.data)
        const backupID = `bak-${Date.now()}`
        const taskID = `backup-create-${Date.now()}`
        mockBackups = [
          {
            backup_id: backupID,
            label: String(body?.label || ''),
            size_bytes: Math.floor(120000 + Math.random() * 180000),
            sha256: `mock-sha-${backupID}`,
            created_at: new Date().toISOString()
          },
          ...mockBackups
        ]
        mockTasks = [
          {
            task_id: taskID,
            task_type: 'backup.create',
            status: 'SUCCEEDED',
            request_json: JSON.stringify(body || {}),
            exit_code: 0,
            stdout_tail: `backup_id=${backupID}`,
            stderr_tail: '',
            log_path: '',
            created_by: mockUser.user_id,
            created_at: new Date().toISOString(),
            started_at: new Date().toISOString(),
            finished_at: new Date().toISOString()
          },
          ...mockTasks
        ]
        return jsonResponse(config, { task_id: taskID, backup_id: backupID, status: 'PENDING' }, 202) as any
      }

      const downloadMatch = url.match(/^\/api\/v1\/backups\/([^/]+)\/download$/)
      if (downloadMatch && method === 'get') {
        return jsonResponse(config, `mock backup file content for ${downloadMatch[1]}`) as any
      }

      const detailMatch = url.match(/^\/api\/v1\/backups\/([^/]+)$/)
      if (detailMatch && method === 'get') {
        const backupID = detailMatch[1]
        const item = mockBackups.find((b) => b.backup_id === backupID)
        if (!item) {
          return jsonResponse(config, { error: 'backup not found' }, 404) as any
        }
        return jsonResponse(config, {
          backup_id: item.backup_id,
          label: item.label,
          scope: ['openclaw_json', 'global_skills'],
          paths: ['/home/openclaw/.openclaw/openclaw.json', '/home/openclaw/.openclaw/skills'],
          sha256: item.sha256,
          created_at: item.created_at
        }) as any
      }

      const restoreMatch = url.match(/^\/api\/v1\/backups\/([^/]+)\/restore$/)
      if (restoreMatch && method === 'post') {
        const backupID = restoreMatch[1]
        const body = parseRequestData(config.data)
        if (body?.dry_run === false) {
          const taskID = `backup-restore-${Date.now()}`
          mockTasks = [
            {
              task_id: taskID,
              task_type: 'backup.restore',
              status: 'SUCCEEDED',
              request_json: JSON.stringify(body || {}),
              exit_code: 0,
              stdout_tail: `restored backup_id=${backupID}`,
              stderr_tail: '',
              log_path: '',
              created_by: mockUser.user_id,
              created_at: new Date().toISOString(),
              started_at: new Date().toISOString(),
              finished_at: new Date().toISOString()
            },
            ...mockTasks
          ]
          return jsonResponse(config, { task_id: taskID, task_type: 'backup.restore', status: 'PENDING' }, 202) as any
        }
        return jsonResponse(config, {
          backup_id: backupID,
          dry_run: true,
          will_overwrite: [
            '/home/openclaw/.openclaw/openclaw.json',
            '/home/openclaw/.openclaw/skills/code-reviewer/README.md'
          ]
        }) as any
      }

      const deleteMatch = url.match(/^\/api\/v1\/backups\/([^/]+)$/)
      if (deleteMatch && method === 'delete') {
        const backupID = deleteMatch[1]
        const before = mockBackups.length
        mockBackups = mockBackups.filter((b) => b.backup_id !== backupID)
        if (mockBackups.length === before) {
          return jsonResponse(config, { error: 'backup not found' }, 404) as any
        }
        return jsonResponse(config, { message: 'deleted' }) as any
      }

      return jsonResponse(config, { error: 'mock route not found' }, 404) as any
    }

    return config
  })
}
