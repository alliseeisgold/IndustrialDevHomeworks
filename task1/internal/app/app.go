package app

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"industrialdev/task1/internal/config"
)

const logFilePath = "/app/logs/app.log"

type App struct {
	cfg    config.Config
	logger *log.Logger
	mu     sync.Mutex
	mux    *http.ServeMux
}

type logRequest struct {
	Message string `json:"message"`
}

func New(cfg config.Config, logger *log.Logger) (*App, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	if err := ensureLogFile(logFilePath); err != nil {
		return nil, err
	}

	application := &App{
		cfg:    cfg,
		logger: logger,
		mux:    http.NewServeMux(),
	}

	application.registerRoutes()

	return application, nil
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) registerRoutes() {
	a.mux.HandleFunc("/", a.handleRoot)
	a.mux.HandleFunc("/status", a.handleStatus)
	a.mux.HandleFunc("/log", a.handleLog)
	a.mux.HandleFunc("/logs", a.handleLogs)
}

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	currentConfig := config.Load()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(currentConfig.WelcomeMessage))
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	defer r.Body.Close()

	var request logRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if strings.TrimSpace(request.Message) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}

	if err := a.appendLog(request.Message); err != nil {
		a.logger.Printf("append log: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to write log"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "logged"})
}

func (a *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			content = []byte{}
		} else {
			a.logger.Printf("read logs: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read logs"})
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}

func (a *App) appendLog(message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(message + "\n"); err != nil {
		return err
	}

	a.logger.Printf("log entry written")

	return nil
}

func ensureLogFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE, 0o644)
	if err != nil {
		return err
	}

	return file.Close()
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
