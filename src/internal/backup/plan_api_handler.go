package backup

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"openclaw-manager/internal/middleware"
)

type createPlanReq struct {
	Name            string   `json:"name"`
	Label           string   `json:"label"`
	Scope           []string `json:"scope"`
	ScheduleKind    string   `json:"schedule_kind"`
	DailyTime       string   `json:"daily_time"`
	MonthlyDay      int      `json:"monthly_day"`
	IntervalMinutes int      `json:"interval_minutes"`
	RetentionCount  int      `json:"retention_count"`
}

func (h *APIHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	if h.PlanSvc == nil {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeInternalError, Message: "plan service unavailable", StatusCode: http.StatusInternalServerError})
		return
	}
	plans, err := h.PlanSvc.ListPlans()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"plans": plans})
}

func (h *APIHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req createPlanReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"name": "required"}))
		return
	}
	if len(req.Scope) == 0 {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"scope": "required"}))
		return
	}
	createdBy := currentUserID(r)
	plan, err := h.PlanSvc.CreatePlan(&BackupPlan{
		Name:            strings.TrimSpace(req.Name),
		Label:           strings.TrimSpace(req.Label),
		Scope:           req.Scope,
		ScheduleKind:    req.ScheduleKind,
		DailyTime:       strings.TrimSpace(req.DailyTime),
		MonthlyDay:      req.MonthlyDay,
		IntervalMinutes: req.IntervalMinutes,
		RetentionCount:  req.RetentionCount,
	}, createdBy)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(plan)
}

func (h *APIHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	planID := planIDFromPath(r.URL.Path)
	if planID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"plan_id": "required"}))
		return
	}
	var req createPlanReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	plan, err := h.PlanSvc.UpdatePlan(planID, &BackupPlan{
		Name:            strings.TrimSpace(req.Name),
		Label:           strings.TrimSpace(req.Label),
		Scope:           req.Scope,
		ScheduleKind:    req.ScheduleKind,
		DailyTime:       strings.TrimSpace(req.DailyTime),
		MonthlyDay:      req.MonthlyDay,
		IntervalMinutes: req.IntervalMinutes,
		RetentionCount:  req.RetentionCount,
	})
	if err == sql.ErrNoRows {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeNotFound, Message: "plan not found", StatusCode: http.StatusNotFound})
		return
	}
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(plan)
}

func (h *APIHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	planID := planIDFromPath(r.URL.Path)
	if err := h.PlanSvc.DeletePlan(planID); err == sql.ErrNoRows {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeNotFound, Message: "plan not found", StatusCode: http.StatusNotFound})
		return
	} else if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"deleted"}`))
}

func (h *APIHandler) EnablePlan(w http.ResponseWriter, r *http.Request) {
	h.togglePlan(w, r, true)
}

func (h *APIHandler) DisablePlan(w http.ResponseWriter, r *http.Request) {
	h.togglePlan(w, r, false)
}

func (h *APIHandler) togglePlan(w http.ResponseWriter, r *http.Request, enabled bool) {
	planID := planIDFromPath(r.URL.Path)
	if err := h.PlanSvc.SetPlanEnabled(planID, enabled); err == sql.ErrNoRows {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeNotFound, Message: "plan not found", StatusCode: http.StatusNotFound})
		return
	} else if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"ok"}`))
}

func (h *APIHandler) RunPlanNow(w http.ResponseWriter, r *http.Request) {
	planID := planIDFromRunPath(r.URL.Path)
	if planID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"plan_id": "required"}))
		return
	}
	plan, err := h.PlanSvc.GetPlan(planID)
	if err == sql.ErrNoRows {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeNotFound, Message: "plan not found", StatusCode: http.StatusNotFound})
		return
	}
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	taskID := h.startBackupTask(plan.Label, plan.Scope, currentUserID(r), planID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"task_id": taskID, "status": "PENDING"})
}

func (h *APIHandler) GetMyPreference(w http.ResponseWriter, r *http.Request) {
	pref, err := h.PlanSvc.GetPreference(currentUserID(r))
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pref)
}

func planIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "backup-plans" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func planIDFromRunPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "backup-plans" && parts[i+2] == "run" {
			return parts[i+1]
		}
	}
	return ""
}
