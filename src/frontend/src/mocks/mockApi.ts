import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { gatewayState, usersMock, tasksMock } from './pageMocks'

const mockToken = 'mock.jwt.token'
const mockUser = { user_id: 'u-admin-001', username: 'admin', role: 'Admin' as const }

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
      if (url.startsWith('/api/v1/users') && method === 'get') {
        return jsonResponse(config, usersMock) as any
      }

      // tasks
      if (url.startsWith('/api/v1/tasks') && method === 'get') {
        return jsonResponse(config, tasksMock) as any
      }

      return jsonResponse(config, { error: 'mock route not found' }, 404) as any
    }

    return config
  })
}
