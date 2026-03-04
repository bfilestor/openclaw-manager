你是一个Go 1.22+ Vue 3 + Element Plus 程序开发专家
将要使用以下技术方案
构建一个运行于同一服务器的轻量 Web 管理平台，提供以下核心能力：

1.  配置文件可视化编辑（带版本历史与原子写入）
2.  Gateway 生命周期管理（启停/重启/状态监控）
3.  多 Agent 状态管理与 Channel Binding 配置
4.  Skills 全局与 Agent 粒度的安装与删除
5.  文件目录备份与还原
6.  实时日志查看与任务状态追踪
7.  （v1.1 新增）用户注册与登录、基于角色的操作权限控制

### 1.1 技术选型
| **层次** | **技术选型**          | **说明**                                            |
|----------|-----------------------|-----------------------------------------------------|
| 后端     | Go 1.22+              | 单二进制，以 mixi 用户运行                          |
| 前端     | Vue 3 + Element Plus  | SPA，由 managerd 静态托管                           |
| 持久化   | SQLite（go-sqlite3）  | 任务队列、Revision、备份元数据、用户表（v1.1）      |
| 认证     | JWT（RS256 或 HS256） | 登录后签发 Token，15min 有效期 + RefreshToken（7d） |
| 密码存储 | bcrypt（cost=12）     | 密码不可逆哈希存储，禁止明文                        |
| 实时通信 | SSE / WebSocket       | 日志流与任务状态推送                                |
| 部署     | systemd user service  | 无需 sudo（MVP 阶段）  

### 1.2 开发测试步骤
0. 先执行git pull 获取最新代码 
1. 再阅读doc/openclaw_requirements_v1.1.md，了解项目整体框架，开发需求
2. 然后阅读doc/todo-list.md，按照依赖顺序，查找可以进行开发的issue
3. 根据issue编号，在doc/dev-plan.md，查找对应的供能描述和测试用例,如果功能需求不明确，可参考doc/openclaw_requirements_v1.1.md，找到对应issue描述，了解具体功能需求
4. 根据功能描述，测试用例进行功能编写及测试用例编写
5. 使用测试用例对编写的功能进行测试，如果测试不通过，重新修改实现功能，只到完全测试通过。
5. 如果功能编写完成并测试通过，在progress.md中编写简要的总结文档，描述完成哪个issue，具体实现什么功能，涉及哪些文件改动，采用以下格式
   - 完成 Issue-01
   -- 功能总结
   Issue-01，完成基础框架搭建
   -- 涉及文件
   src/front/index.vue,
   src/backend/test.go
6. 更新doc/todo-list.md对应issue状态及测试状态。
7. 提交git保存版本，备注：完成Issue-xxx,简要描述功能，并查找下一个开发Issue，进行开发测试，只到全部issue开发测试完成。
### 1.3 开发要求
- 所有代码都必须放置在src目录下，前端代码放置在src/front目录下，后端代码放在 src/backend目录下
- 要有足够中文备注信息
- 功能尽量内聚，并可重复使用
- 开发新的issue时非必要，不要修改前面实现的功能
- 如果请求AI接口出现429或者超时，则暂停编写，并告知用户
				
