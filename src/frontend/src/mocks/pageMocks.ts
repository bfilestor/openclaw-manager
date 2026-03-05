export type Role = 'Viewer' | 'Operator' | 'Admin'

export type GatewayState = {
  service: {
    active_state: 'active' | 'inactive'
    sub_state: 'running' | 'dead'
    main_pid: string
    exec_start: string
    fragment_path: string
    active_enter_timestamp: string
  }
  bind_addr: string
  port: number
  log_path: string
  node_path: string
  nvm_warning: boolean
}

export const gatewayState: GatewayState = {
  service: {
    active_state: 'active',
    sub_state: 'running',
    main_pid: '12345',
    exec_start: '/usr/bin/openclaw gateway start',
    fragment_path: '/home/mixi/.config/systemd/user/openclaw-gateway.service',
    active_enter_timestamp: new Date().toISOString()
  },
  bind_addr: '127.0.0.1',
  port: 18790,
  log_path: '/tmp/openclaw/openclaw-2026-03-05.log',
  node_path: '/home/mixi/.nvm/versions/node/v24.14.0/bin/node',
  nvm_warning: false
}

export const usersMock = {
  users: [
    {
      user_id: 'u-admin-001',
      username: 'admin',
      role: 'Admin' as Role,
      status: 'active',
      created_at: new Date().toISOString(),
      last_login_at: new Date().toISOString()
    },
    {
      user_id: 'u-op-001',
      username: 'operator',
      role: 'Operator' as Role,
      status: 'active',
      created_at: new Date().toISOString()
    },
    {
      user_id: 'u-view-001',
      username: 'viewer',
      role: 'Viewer' as Role,
      status: 'disabled',
      created_at: new Date().toISOString()
    }
  ],
  total: 3
}

export const tasksMock = {
  tasks: [
    {
      task_id: 'task-001',
      task_type: 'gateway.restart',
      status: 'SUCCEEDED',
      request_json: '{"action":"restart"}',
      exit_code: 0,
      stdout_tail: 'gateway restarted',
      stderr_tail: '',
      log_path: '/tmp/openclaw/task-001.log',
      created_by: 'u-admin-001',
      created_at: new Date().toISOString(),
      started_at: new Date().toISOString(),
      finished_at: new Date().toISOString()
    },
    {
      task_id: 'task-002',
      task_type: 'backup.create',
      status: 'PENDING',
      request_json: '{"scope":["openclaw_json"]}',
      exit_code: null,
      stdout_tail: '',
      stderr_tail: '',
      log_path: '',
      created_by: 'u-op-001',
      created_at: new Date().toISOString(),
      started_at: null,
      finished_at: null
    }
  ],
  total: 2
}
