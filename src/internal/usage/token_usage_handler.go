package usage

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/user"
)

const (
	maxConversationRecords = 100
	defaultPageSize        = 20
	maxPageSize            = 50
)

type TokenUsageHandler struct {
	OpenClawHome string
	AccountBinds *auth.AccountBindingRepository
}

type providerCostIndex map[string]float64

type sessionMeta struct {
	AgentID       string
	SessionKey    string
	SessionID     string
	UpdatedAt     int64
	ModelProvider string
	Model         string
	AccountID     string
	InputTokens   int64
	OutputTokens  int64
	TotalTokens   int64
	SessionFile   string
}

type botSummary struct {
	BotID         string  `json:"botId"`
	Sessions      int     `json:"sessions"`
	InputTokens   int64   `json:"inputTokens"`
	OutputTokens  int64   `json:"outputTokens"`
	TotalTokens   int64   `json:"totalTokens"`
	EstimatedCost float64 `json:"estimatedCost"`
}

type conversationRow struct {
	SessionKey    string  `json:"sessionKey"`
	SessionID     string  `json:"sessionId"`
	AgentID       string  `json:"agentId"`
	UpdatedAt     string  `json:"updatedAt"`
	ModelProvider string  `json:"modelProvider"`
	Model         string  `json:"model"`
	InputTokens   int64   `json:"inputTokens"`
	OutputTokens  int64   `json:"outputTokens"`
	TotalTokens   int64   `json:"totalTokens"`
	EstimatedCost float64 `json:"estimatedCost"`
	Preview       string  `json:"preview"`
}

type sessionMessage struct {
	Role      string `json:"role"`
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}

func (h *TokenUsageHandler) Summary(w http.ResponseWriter, r *http.Request) {
	rangeDays := parsePositiveInt(r.URL.Query().Get("days"), 0)
	sessions, costs, err := h.loadAllSessionMeta()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	if rangeDays > 0 {
		sessions = filterSessionsByDays(sessions, rangeDays)
	}
	boundAccountID, err := h.resolveScopeAccountID(r)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if boundAccountID != "" {
		sessions = filterSessionsByAccount(sessions, boundAccountID)
	}

	botMap := map[string]*botSummary{}
	var totalInput, totalOutput, totalTokens int64
	var totalCost float64

	for _, s := range sessions {
		botID := safeBotID(s.AccountID)
		item, ok := botMap[botID]
		if !ok {
			item = &botSummary{BotID: botID}
			botMap[botID] = item
		}
		item.Sessions++
		item.InputTokens += s.InputTokens
		item.OutputTokens += s.OutputTokens
		item.TotalTokens += s.TotalTokens
		cost := estimateCost(s.TotalTokens, costs[s.ModelProvider])
		item.EstimatedCost += cost

		totalInput += s.InputTokens
		totalOutput += s.OutputTokens
		totalTokens += s.TotalTokens
		totalCost += cost
	}

	rows := make([]botSummary, 0, len(botMap))
	for _, v := range botMap {
		rows = append(rows, *v)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].TotalTokens == rows[j].TotalTokens {
			return rows[i].BotID < rows[j].BotID
		}
		return rows[i].TotalTokens > rows[j].TotalTokens
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"days": rangeDays,
		"total": map[string]any{
			"inputTokens":   totalInput,
			"outputTokens":  totalOutput,
			"totalTokens":   totalTokens,
			"estimatedCost": totalCost,
		},
		"bots": rows,
	})
}

func (h *TokenUsageHandler) BotConversations(w http.ResponseWriter, r *http.Request) {
	botID := strings.TrimSpace(r.PathValue("botId"))
	if botID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"botId": "required"}))
		return
	}

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	pageSize := parsePositiveInt(r.URL.Query().Get("page_size"), defaultPageSize)
	rangeDays := parsePositiveInt(r.URL.Query().Get("days"), 0)
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	sessions, costs, err := h.loadAllSessionMeta()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	if rangeDays > 0 {
		sessions = filterSessionsByDays(sessions, rangeDays)
	}
	boundAccountID, err := h.resolveScopeAccountID(r)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}
	if boundAccountID != "" {
		if botID != boundAccountID {
			middleware.WriteAppError(w, middleware.NewForbidden("bound account only"))
			return
		}
		sessions = filterSessionsByAccount(sessions, boundAccountID)
	}

	filtered := make([]sessionMeta, 0)
	for _, s := range sessions {
		if safeBotID(s.AccountID) == botID {
			filtered = append(filtered, s)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].UpdatedAt > filtered[j].UpdatedAt
	})

	if len(filtered) > maxConversationRecords {
		filtered = filtered[:maxConversationRecords]
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	rows := make([]conversationRow, 0, end-start)
	for _, s := range filtered[start:end] {
		rows = append(rows, conversationRow{
			SessionKey:    s.SessionKey,
			SessionID:     s.SessionID,
			AgentID:       s.AgentID,
			UpdatedAt:     msToRFC3339(s.UpdatedAt),
			ModelProvider: s.ModelProvider,
			Model:         s.Model,
			InputTokens:   s.InputTokens,
			OutputTokens:  s.OutputTokens,
			TotalTokens:   s.TotalTokens,
			EstimatedCost: estimateCost(s.TotalTokens, costs[s.ModelProvider]),
			Preview:       readSessionPreview(s.SessionFile),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"botId":    botID,
		"days":     rangeDays,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"items":    rows,
	})
}

func (h *TokenUsageHandler) SessionMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.PathValue("sessionId"))
	if sessionID == "" {
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"sessionId": "required"}))
		return
	}
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 80)
	if limit > 200 {
		limit = 200
	}

	sessions, _, err := h.loadAllSessionMeta()
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	boundAccountID, err := h.resolveScopeAccountID(r)
	if err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	targetFile := ""
	targetAccountID := ""
	for _, s := range sessions {
		if s.SessionID == sessionID {
			targetFile = s.SessionFile
			targetAccountID = safeBotID(s.AccountID)
			break
		}
	}
	if boundAccountID != "" && boundAccountID != targetAccountID {
		middleware.WriteAppError(w, middleware.NewForbidden("bound account only"))
		return
	}
	if targetFile == "" {
		middleware.WriteAppError(w, &middleware.AppError{Code: middleware.CodeNotFound, Message: "session not found", StatusCode: http.StatusNotFound})
		return
	}

	messages := readSessionMessages(targetFile, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"sessionId": sessionID,
		"count":     len(messages),
		"items":     messages,
	})
}

func (h *TokenUsageHandler) loadAllSessionMeta() ([]sessionMeta, providerCostIndex, error) {
	openclawHome := strings.TrimSpace(h.OpenClawHome)
	if openclawHome == "" {
		return nil, nil, errors.New("openclaw home is empty")
	}

	costs := loadProviderCosts(filepath.Join(openclawHome, "openclaw.json"))
	agentsRoot := filepath.Join(openclawHome, "agents")
	entries, err := os.ReadDir(agentsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []sessionMeta{}, costs, nil
		}
		return nil, nil, err
	}

	out := make([]sessionMeta, 0, 128)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentID := entry.Name()
		storePath := filepath.Join(agentsRoot, agentID, "sessions", "sessions.json")
		rows, err := loadAgentSessions(storePath, agentID)
		if err != nil {
			continue
		}
		out = append(out, rows...)
	}
	return out, costs, nil
}

func loadProviderCosts(path string) providerCostIndex {
	raw, err := os.ReadFile(path)
	if err != nil {
		return providerCostIndex{}
	}
	var cfg struct {
		Models struct {
			Providers map[string]struct {
				XManager struct {
					CostPer1KToken float64 `json:"costPer1kToken"`
				} `json:"xManager"`
			} `json:"providers"`
		} `json:"models"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return providerCostIndex{}
	}

	out := providerCostIndex{}
	for provider, item := range cfg.Models.Providers {
		if item.XManager.CostPer1KToken < 0 {
			continue
		}
		out[provider] = item.XManager.CostPer1KToken
	}
	return out
}

func loadAgentSessions(path, agentID string) ([]sessionMeta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	store := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &store); err != nil {
		return nil, err
	}

	rows := make([]sessionMeta, 0, len(store))
	for key, item := range store {
		var rec struct {
			SessionID     string `json:"sessionId"`
			UpdatedAt     int64  `json:"updatedAt"`
			ModelProvider string `json:"modelProvider"`
			Model         string `json:"model"`
			InputTokens   int64  `json:"inputTokens"`
			OutputTokens  int64  `json:"outputTokens"`
			TotalTokens   int64  `json:"totalTokens"`
			SessionFile   string `json:"sessionFile"`
			Origin        struct {
				AccountID string `json:"accountId"`
			} `json:"origin"`
			DeliveryContext struct {
				AccountID string `json:"accountId"`
			} `json:"deliveryContext"`
			LastAccountID string `json:"lastAccountId"`
		}
		if err := json.Unmarshal(item, &rec); err != nil {
			continue
		}

		accountID := firstNonEmpty(rec.LastAccountID, rec.DeliveryContext.AccountID, rec.Origin.AccountID)
		if accountID == "" {
			accountID = deriveBotIDFromSessionKey(key)
		}

		rows = append(rows, sessionMeta{
			AgentID:       agentID,
			SessionKey:    key,
			SessionID:     rec.SessionID,
			UpdatedAt:     rec.UpdatedAt,
			ModelProvider: rec.ModelProvider,
			Model:         rec.Model,
			AccountID:     accountID,
			InputTokens:   rec.InputTokens,
			OutputTokens:  rec.OutputTokens,
			TotalTokens:   rec.TotalTokens,
			SessionFile:   rec.SessionFile,
		})
	}
	return rows, nil
}

func deriveBotIDFromSessionKey(sessionKey string) string {
	parts := strings.Split(sessionKey, ":")
	if len(parts) >= 3 {
		return parts[1]
	}
	return "unknown"
}

func safeBotID(botID string) string {
	botID = strings.TrimSpace(botID)
	if botID == "" {
		return "default"
	}
	return botID
}

func (h *TokenUsageHandler) resolveScopeAccountID(r *http.Request) (string, error) {
	uc, ok := auth.GetUserContext(r.Context())
	if !ok || uc == nil {
		return "", middleware.NewUnauthorized()
	}
	if uc.Role != user.RoleUser {
		return "", nil
	}
	if h.AccountBinds == nil {
		return "", &middleware.AppError{Code: middleware.CodePermissionDenied, Message: "user role requires account binding", StatusCode: http.StatusForbidden}
	}
	item, err := h.AccountBinds.GetByUserID(uc.UserID)
	if err != nil {
		if errors.Is(err, auth.ErrAccountBindingNotFound) {
			return "", &middleware.AppError{Code: middleware.CodePermissionDenied, Message: "user role requires account binding", StatusCode: http.StatusForbidden}
		}
		return "", err
	}
	return safeBotID(item.AccountID), nil
}

func filterSessionsByDays(sessions []sessionMeta, days int) []sessionMeta {
	if days <= 0 {
		return sessions
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).UnixMilli()
	filtered := make([]sessionMeta, 0, len(sessions))
	for _, s := range sessions {
		if s.UpdatedAt >= cutoff {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func filterSessionsByAccount(sessions []sessionMeta, accountID string) []sessionMeta {
	accountID = safeBotID(accountID)
	out := make([]sessionMeta, 0, len(sessions))
	for _, s := range sessions {
		if safeBotID(s.AccountID) == accountID {
			out = append(out, s)
		}
	}
	return out
}

func estimateCost(totalTokens int64, costPer1k float64) float64 {
	if totalTokens <= 0 || costPer1k <= 0 {
		return 0
	}
	return (float64(totalTokens) / 1000.0) * costPer1k
}

func parsePositiveInt(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func readSessionPreview(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	line := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()
	}
	if line == "" {
		return ""
	}

	var msg struct {
		Role    string `json:"role"`
		Message string `json:"message"`
		Text    string `json:"text"`
		Content any    `json:"content"`
	}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return ""
	}
	preview := firstNonEmpty(msg.Message, msg.Text, stringifyContent(msg.Content))
	preview = strings.TrimSpace(preview)
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}
	return preview
}

func stringifyContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if t, ok := m["text"].(string); ok && strings.TrimSpace(t) != "" {
				parts = append(parts, t)
			}
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}

func readSessionMessages(path string, limit int) []sessionMessage {
	path = strings.TrimSpace(path)
	if path == "" {
		return []sessionMessage{}
	}
	file, err := os.Open(path)
	if err != nil {
		return []sessionMessage{}
	}
	defer file.Close()

	rows := make([]sessionMessage, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		msg, ok := parseJSONLMessage(line)
		if !ok {
			continue
		}
		rows = append(rows, msg)
	}
	if len(rows) > limit {
		rows = rows[len(rows)-limit:]
	}
	return rows
}

func parseJSONLMessage(line string) (sessionMessage, bool) {
	var row struct {
		Role      string `json:"role"`
		Timestamp int64  `json:"timestamp"`
		CreatedAt int64  `json:"createdAt"`
		Message   string `json:"message"`
		Text      string `json:"text"`
		Content   any    `json:"content"`
	}
	if err := json.Unmarshal([]byte(line), &row); err != nil {
		return sessionMessage{}, false
	}
	text := firstNonEmpty(row.Message, row.Text, stringifyContent(row.Content))
	if strings.TrimSpace(text) == "" {
		return sessionMessage{}, false
	}
	ts := row.Timestamp
	if ts <= 0 {
		ts = row.CreatedAt
	}
	return sessionMessage{Role: firstNonEmpty(row.Role, "unknown"), Timestamp: msToRFC3339(ts), Text: text}, true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func msToRFC3339(ts int64) string {
	if ts <= 0 {
		return ""
	}
	sec := ts / 1000
	nsec := (ts % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec).UTC().Format(time.RFC3339)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
