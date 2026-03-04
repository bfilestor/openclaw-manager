package gateway

import (
	"context"
	"net/http"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
)

type DoctorHandler struct {
	Exec    Executor
	Timeout time.Duration
}

func NewDoctorHandler(exec Executor) *DoctorHandler {
	if exec == nil {
		exec = OSExecutor{}
	}
	return &DoctorHandler{Exec: exec, Timeout: 5 * time.Minute}
}

func (h *DoctorHandler) Run(w http.ResponseWriter, r *http.Request) {
	h.execDoctor(w, false)
}

func (h *DoctorHandler) Repair(w http.ResponseWriter, r *http.Request) {
	h.execDoctor(w, true)
}

func (h *DoctorHandler) execDoctor(w http.ResponseWriter, repair bool) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()
	args := []string{"doctor"}
	if repair {
		args = append(args, "--repair")
	}
	out, err := h.Exec.Run(ctx, "openclaw", args...)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			middleware.WriteAppError(w, &middleware.AppError{Code: "TIMEOUT", Message: "doctor timeout", StatusCode: http.StatusGatewayTimeout})
			return
		}
		middleware.WriteAppError(w, err)
		return
	}
	output := string(out)
	nvmDetected := strings.Contains(strings.ToLower(output), ".nvm")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if repair {
		_, _ = w.Write([]byte(`{"task_type":"doctor.repair","status":"PENDING","nvm_detected":` + boolJSON(nvmDetected) + `}`))
		return
	}
	_, _ = w.Write([]byte(`{"task_type":"doctor.run","status":"PENDING","nvm_detected":` + boolJSON(nvmDetected) + `}`))
}

func boolJSON(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
