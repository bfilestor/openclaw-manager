package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"openclaw-manager/internal/agent"
	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/backup"
	appcfg "openclaw-manager/internal/config"
	"openclaw-manager/internal/gateway"
	"openclaw-manager/internal/server"
	"openclaw-manager/internal/skills"
	"openclaw-manager/internal/storage"
	"openclaw-manager/internal/task"
	"openclaw-manager/internal/usage"
	"openclaw-manager/internal/user"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config.toml")
	staticDir := flag.String("static-dir", "", "path to frontend dist directory")
	flag.Parse()

	if err := validateConfigPath(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := appcfg.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config error: %v\n", err)
		os.Exit(1)
	}

	dbPath := resolveDBPath(cfg)
	db, err := storage.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db error: %v\n", err)
		os.Exit(1)
	}
	defer db.SQL.Close()

	passSvc := auth.NewPasswordService()
	if cfg.Auth.PasswordMinLen > 0 {
		passSvc.MinLength = cfg.Auth.PasswordMinLen
	}
	tokenRepo := auth.NewTokenRepository(db.SQL)
	jwtSvc := &auth.JWTService{
		Secret:           []byte(cfg.Auth.JWTSecret),
		AccessTokenTTL:   cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL:  cfg.Auth.RefreshTokenTTL,
		BlacklistChecker: tokenRepo,
	}
	accountBindRepo := auth.NewAccountBindingRepository(db.SQL)
	systemSettingsRepo := auth.NewSystemSettingsRepository(db.SQL)
	authHandler := &auth.Handler{
		Repo:         user.NewRepository(db.SQL),
		Pass:         passSvc,
		Config:       cfg,
		JWT:          jwtSvc,
		TokenRepo:    tokenRepo,
		AccountBinds: accountBindRepo,
		Settings:     systemSettingsRepo,
	}

	dist := resolveStaticDir(*staticDir)
	s := server.New(cfg.Server.Listen, dist, registerAllRoutes(cfg, db.SQL, authHandler, jwtSvc))
	fmt.Printf("manager server starting, listen=%s, static_dir=%s, db=%s\n", cfg.Server.Listen, dist, dbPath)

	if err := server.RunWithSignals(s); err != nil {
		fmt.Fprintf(os.Stderr, "server run error: %v\n", err)
		os.Exit(1)
	}
}

func registerAllRoutes(cfg *appcfg.Config, sqlDB *sql.DB, authHandler *auth.Handler, jwtSvc *auth.JWTService) func(*http.ServeMux) {
	return func(mux *http.ServeMux) {
		authMW := auth.AuthMiddleware(jwtSvc)
		wrap := func(hf http.HandlerFunc, mws ...func(http.Handler) http.Handler) http.HandlerFunc {
			h := http.Handler(http.HandlerFunc(hf))
			for i := len(mws) - 1; i >= 0; i-- {
				h = mws[i](h)
			}
			return h.ServeHTTP
		}

		// 基础依赖
		home, _ := os.UserHomeDir()
		validator, _ := storage.NewPathValidator([]string{
			cfg.Paths.OpenClawHome,
			cfg.Paths.ManagerHome,
			filepath.Join(home, ".config", "systemd", "user"),
			"/tmp/openclaw",
		})
		execer := gateway.OSExecutor{}
		revRepo := appcfg.NewRevisionRepository(sqlDB)
		agentRepo := agent.NewRepository(execer, validator)
		taskRepo := task.NewRepository(sqlDB)
		accountBindRepo := auth.NewAccountBindingRepository(sqlDB)

		// 各功能 handler
		gatewaySvc := gateway.NewSystemctlService(execer)
		gatewayAPI := &gateway.APIHandler{Service: gatewaySvc, ServiceName: "openclaw-gateway.service"}
		gatewayDoctor := gateway.NewDoctorHandler(execer)
		gatewayLogs := gateway.NewLogsHandler(execer)

		agentAPI := &agent.Handler{Repo: agentRepo}
		agentSessionAPI := &agent.SessionStatusHandler{Exec: execer}
		agentManage := &agent.ManageHandler{
			Exec:             execer,
			Workspaces:       agentRepo,
			OpenClawJSONPath: filepath.Join(cfg.Paths.OpenClawHome, "openclaw.json"),
		}
		bindingAPI := agent.NewBindingHandler(execer)

		skillList := &skills.Handler{AgentRepo: agentRepo, GlobalDir: filepath.Join(cfg.Paths.OpenClawHome, "skills")}
		skillInstall := &skills.InstallHandler{GlobalDir: filepath.Join(cfg.Paths.OpenClawHome, "skills"), AgentRepo: agentRepo, Validator: validator}
		skillDelete := &skills.DeleteHandler{AgentRepo: agentRepo, GlobalDir: filepath.Join(cfg.Paths.OpenClawHome, "skills"), Validator: validator}

		backupSvc := &backup.Service{
			DB:           sqlDB,
			BackupHome:   filepath.Join(cfg.Paths.ManagerHome, "backups"),
			OpenclawHome: cfg.Paths.OpenClawHome,
			ManagerHome:  cfg.Paths.ManagerHome,
		}
		backupPlanSvc := &backup.PlanService{DB: sqlDB, Backup: backupSvc}
		backupPlanSvc.Start()
		backupAPI := &backup.APIHandler{Service: backupSvc, PlanSvc: backupPlanSvc, DB: sqlDB}

		taskAPI := &task.Handler{Repo: taskRepo}
		taskSSE := &task.SSEHandler{Repo: taskRepo}
		taskShell := task.NewShellHandler(execer)

		openclawJSON := &appcfg.OpenClawJSONHandler{
			FilePath:  filepath.Join(cfg.Paths.OpenClawHome, "openclaw.json"),
			Validator: validator,
			Revisions: revRepo,
		}
		identityAPI := &appcfg.IdentityHandler{AgentRepo: agentRepo, Revisions: revRepo, Validator: validator}
		tokenUsageAPI := &usage.TokenUsageHandler{OpenClawHome: cfg.Paths.OpenClawHome, AccountBinds: accountBindRepo}

		// 公开接口
		mux.HandleFunc("GET /api/v1/auth/public-registration", authHandler.PublicRegistrationStatus)
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
		mux.HandleFunc("GET /api/v1/auth/resetpwd/admin", authHandler.GetResetPasswordAdmin)
		mux.HandleFunc("POST /api/v1/auth/resetpwd", authHandler.ResetFirstAdminPassword)

		// 需要认证的接口
		mux.HandleFunc("POST /api/v1/auth/logout", wrap(authHandler.Logout, authMW))
		mux.HandleFunc("GET /api/v1/users/me", wrap(authHandler.Me, authMW))
		mux.HandleFunc("PUT /api/v1/users/me/password", wrap(authHandler.ChangeMyPassword, authMW))
		mux.HandleFunc("GET /api/v1/system/settings", wrap(authHandler.GetSystemSettings, authMW))
		mux.HandleFunc("PUT /api/v1/system/settings", wrap(authHandler.PutSystemSettings, authMW))
		mux.HandleFunc("GET /api/v1/users", wrap(authHandler.ListUsers, authMW))
		mux.HandleFunc("POST /api/v1/users", wrap(authHandler.CreateUser, authMW))
		mux.HandleFunc("PUT /api/v1/users/{id}/role", wrap(authHandler.UpdateUserRole, authMW))
		mux.HandleFunc("PUT /api/v1/users/{id}/password", wrap(authHandler.UpdateUserPassword, authMW))
		mux.HandleFunc("GET /api/v1/users/me/account-binding", wrap(authHandler.GetMyAccountBinding, authMW))
		mux.HandleFunc("GET /api/v1/users/{id}/account-binding", wrap(authHandler.GetUserAccountBinding, authMW))
		mux.HandleFunc("PUT /api/v1/users/{id}/account-binding", wrap(authHandler.SetUserAccountBinding, authMW))
		mux.HandleFunc("POST /api/v1/users/{id}/disable", wrap(authHandler.DisableUser, authMW))
		mux.HandleFunc("DELETE /api/v1/users/{id}", wrap(authHandler.DeleteUser, authMW))

		mux.HandleFunc("GET /api/v1/tasks", wrap(taskAPI.ListTasks, authMW))
		mux.HandleFunc("DELETE /api/v1/tasks", wrap(taskAPI.ClearTasks, authMW))
		mux.HandleFunc("GET /api/v1/tasks/{id}", wrap(taskAPI.GetTask, authMW))
		mux.HandleFunc("DELETE /api/v1/tasks/{id}", wrap(taskAPI.DeleteTask, authMW))
		mux.HandleFunc("POST /api/v1/tasks/{id}/cancel", wrap(taskAPI.CancelTask, authMW))
		mux.HandleFunc("GET /api/v1/tasks/{id}/events", wrap(taskSSE.TaskEvents, authMW))
		mux.HandleFunc("POST /api/v1/tasks/shell/execute", wrap(taskShell.Execute, authMW, auth.RequireRole(user.RoleOperator)))

		mux.HandleFunc("GET /api/v1/gateway/status", wrap(gatewayAPI.Status, authMW))
		mux.HandleFunc("POST /api/v1/gateway/start", wrap(gatewayAPI.Start, authMW))
		mux.HandleFunc("POST /api/v1/gateway/stop", wrap(gatewayAPI.Stop, authMW))
		mux.HandleFunc("POST /api/v1/gateway/restart", wrap(gatewayAPI.Restart, authMW))
		mux.HandleFunc("GET /api/v1/gateway/logs", wrap(gatewayLogs.GetLogs, authMW))
		mux.HandleFunc("POST /api/v1/gateway/doctor", wrap(gatewayDoctor.Run, authMW))
		mux.HandleFunc("POST /api/v1/gateway/doctor/repair", wrap(gatewayDoctor.Repair, authMW))

		mux.HandleFunc("GET /api/v1/agents", wrap(agentAPI.ListAgents, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}", wrap(agentAPI.GetAgent, authMW))
		mux.HandleFunc("GET /api/v1/agent-sessions", wrap(agentSessionAPI.ListAgentSessions, authMW))
		mux.HandleFunc("POST /api/v1/agents", wrap(agentManage.CreateAgent, authMW))
		mux.HandleFunc("DELETE /api/v1/agents/{id}", wrap(agentManage.DeleteAgent, authMW))
		workspaceMD := &appcfg.WorkspaceMarkdownHandler{Workspaces: agentRepo, Validator: validator, Revisions: revRepo}
		mux.HandleFunc("POST /api/v1/agents/{id}/workspace/migrate", wrap(agentManage.MigrateWorkspace, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}/workspace/markdown/files", wrap(workspaceMD.ListFiles, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}/workspace/markdown/file", wrap(workspaceMD.GetFile, authMW))
		mux.HandleFunc("PUT /api/v1/agents/{id}/workspace/markdown/file", wrap(workspaceMD.PutFile, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}/workspace/markdown/revisions", wrap(workspaceMD.ListRevisions, authMW))
		mux.HandleFunc("POST /api/v1/agents/{id}/workspace/markdown/revisions/{rev_id}/restore", wrap(workspaceMD.RestoreRevision, authMW))
		mux.HandleFunc("DELETE /api/v1/agents/{id}/workspace/markdown/revisions/{rev_id}", wrap(workspaceMD.DeleteRevision, authMW))
		mux.HandleFunc("GET /api/v1/bindings", wrap(bindingAPI.ListBindings, authMW))
		mux.HandleFunc("POST /api/v1/bindings/apply", wrap(bindingAPI.ApplyBindings, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}/identity", wrap(identityAPI.GetIdentity, authMW))
		mux.HandleFunc("PUT /api/v1/agents/{id}/identity", wrap(identityAPI.PutIdentity, authMW))
		mux.HandleFunc("GET /api/v1/agents/{id}/identity/revisions", wrap(identityAPI.ListIdentityRevisions, authMW))

		mux.HandleFunc("GET /api/v1/skills", wrap(skillList.ListSkills, authMW))
		mux.HandleFunc("POST /api/v1/skills/install", wrap(skillInstall.InstallSkill, authMW))
		mux.HandleFunc("DELETE /api/v1/skills/{name}", wrap(skillDelete.DeleteSkill, authMW))

		mux.HandleFunc("GET /api/v1/token-usage/summary", wrap(tokenUsageAPI.Summary, authMW))
		mux.HandleFunc("GET /api/v1/token-usage/bots/{botId}/conversations", wrap(tokenUsageAPI.BotConversations, authMW))
		mux.HandleFunc("GET /api/v1/token-usage/sessions/{sessionId}/messages", wrap(tokenUsageAPI.SessionMessages, authMW))

		mux.HandleFunc("GET /api/v1/config/openclaw", wrap(openclawJSON.GetOpenClawJSON, authMW))
		mux.HandleFunc("PUT /api/v1/config/openclaw", wrap(openclawJSON.PutOpenClawJSON, authMW))
		mux.HandleFunc("GET /api/v1/config/openclaw/revisions", wrap(openclawJSON.ListRevisions, authMW))
		mux.HandleFunc("POST /api/v1/config/openclaw/revisions/{id}/restore", wrap(openclawJSON.RestoreRevision, authMW))
		mux.HandleFunc("DELETE /api/v1/config/openclaw/revisions/{id}", wrap(openclawJSON.DeleteRevision, authMW))

		mux.HandleFunc("POST /api/v1/backups", wrap(backupAPI.CreateBackup, authMW))
		mux.HandleFunc("GET /api/v1/backups", wrap(backupAPI.ListBackups, authMW))
		mux.HandleFunc("GET /api/v1/backups/{id}", wrap(backupAPI.GetBackup, authMW))
		mux.HandleFunc("GET /api/v1/backups/{id}/download", wrap(backupAPI.DownloadBackup, authMW))
		mux.HandleFunc("POST /api/v1/backups/{id}/restore", wrap(backupAPI.RestoreBackup, authMW))
		mux.HandleFunc("DELETE /api/v1/backups/{id}", wrap(backupAPI.DeleteBackup, authMW))
		mux.HandleFunc("GET /api/v1/backup-plans", wrap(backupAPI.ListPlans, authMW))
		mux.HandleFunc("POST /api/v1/backup-plans", wrap(backupAPI.CreatePlan, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("PUT /api/v1/backup-plans/{id}", wrap(backupAPI.UpdatePlan, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("DELETE /api/v1/backup-plans/{id}", wrap(backupAPI.DeletePlan, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("POST /api/v1/backup-plans/{id}/enable", wrap(backupAPI.EnablePlan, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("POST /api/v1/backup-plans/{id}/disable", wrap(backupAPI.DisablePlan, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("POST /api/v1/backup-plans/{id}/run", wrap(backupAPI.RunPlanNow, authMW, auth.RequireRole(user.RoleOperator)))
		mux.HandleFunc("GET /api/v1/backup-preferences/me", wrap(backupAPI.GetMyPreference, authMW))
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.openclaw-manager/config.toml"
	}
	return filepath.Join(home, ".openclaw-manager", "config.toml")
}

func validateConfigPath(path string) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func resolveStaticDir(fromFlag string) string {
	if fromFlag != "" {
		return fromFlag
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, "frontend", "dist")
}

func resolveDBPath(cfg *appcfg.Config) string {
	if cfg != nil && cfg.Paths.ManagerHome != "" {
		return filepath.Join(cfg.Paths.ManagerHome, "manager.db")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "manager.db")
	}
	return filepath.Join(home, ".openclaw-manager", "manager.db")
}
