package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flowpilot/flowpilot/internal/domain"
	"github.com/flowpilot/flowpilot/internal/storage"
)

const maxRequestBody = 1 << 20

func main() {
	ctx := context.Background()
	store, err := storage.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ready"}) })
	mux.HandleFunc("/hooks/", webhook(store))
	mux.HandleFunc("/api/v1/workflows", workflows(store))
	mux.HandleFunc("/api/v1/workflows/", workflow(store))
	mux.HandleFunc("/api/v1/executions/", execution(store))

	server := http.Server{
		Addr:              env("FLOWPILOT_API_ADDR", ":8080"),
		Handler:           cors(requestID(limitBody(mux))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("api listening", "addr", server.Addr)
	slog.Error("server stopped", "error", server.ListenAndServe())
}

func workflows(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := store.ListWorkflows(r.Context())
			if err != nil {
				problem(w, http.StatusInternalServerError, "could not list workflows")
				return
			}
			writeJSON(w, http.StatusOK, items)
		case http.MethodPost:
			var input struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Name) == "" {
				problem(w, http.StatusBadRequest, "name is required and body must be valid JSON")
				return
			}
			item, err := store.CreateWorkflow(r.Context(), strings.TrimSpace(input.Name), input.Description)
			if err != nil {
				problem(w, http.StatusInternalServerError, "could not create workflow")
				return
			}
			writeJSON(w, http.StatusCreated, item)
		default:
			problem(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func workflow(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/"), "/")
		if len(parts) < 1 || parts[0] == "" {
			problem(w, http.StatusNotFound, "not found")
			return
		}
		id := parts[0]
		if len(parts) != 2 {
			problem(w, http.StatusNotFound, "not found")
			return
		}

		switch parts[1] {
		case "draft":
			if r.Method != http.MethodPut {
				problem(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var input struct {
				Revision int64        `json:"revision"`
				Graph    domain.Graph `json:"graph"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				problem(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if validation := domain.ValidateGraph(input.Graph); len(validation) > 0 {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": validation})
				return
			}
			item, err := store.UpdateDraft(r.Context(), id, input.Revision, input.Graph)
			if err != nil {
				problem(w, http.StatusConflict, "draft revision is stale or workflow does not exist")
				return
			}
			writeJSON(w, http.StatusOK, item)
		case "publish":
			if r.Method != http.MethodPost {
				problem(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			item, err := store.Publish(r.Context(), id)
			if err != nil {
				problem(w, http.StatusConflict, "workflow cannot be published")
				return
			}
			writeJSON(w, http.StatusOK, item)
		case "executions":
			if r.Method == http.MethodGet {
				items, err := store.ListExecutions(r.Context(), id)
				if err != nil {
					problem(w, http.StatusInternalServerError, "could not list executions")
					return
				}
				writeJSON(w, http.StatusOK, items)
				return
			}
			if r.Method != http.MethodPost {
				problem(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var input struct{ Input map[string]any `json:"input"` }
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
				problem(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			item, err := store.Start(r.Context(), id, input.Input)
			if err != nil {
				problem(w, http.StatusConflict, "workflow is not published or does not exist")
				return
			}
			writeJSON(w, http.StatusAccepted, item)
		default:
			problem(w, http.StatusNotFound, "not found")
		}
	}
}

func execution(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/cancel") {
			problem(w, http.StatusNotFound, "not found")
			return
		}
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/executions/"), "/cancel")
		if id == "" {
			problem(w, http.StatusNotFound, "not found")
			return
		}
		if err := store.Cancel(r.Context(), id); err != nil {
			problem(w, http.StatusConflict, "execution cannot be cancelled")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func webhook(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			problem(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		hook := strings.TrimPrefix(r.URL.Path, "/hooks/")
		if hook == "" || strings.Contains(hook, "/") {
			problem(w, http.StatusNotFound, "webhook not found")
			return
		}
		payload := map[string]any{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			problem(w, http.StatusBadRequest, "body must be JSON under 1 MiB")
			return
		}
		payload["headers"] = r.Header
		payload["query"] = r.URL.Query()
		item, err := store.StartWebhook(r.Context(), hook, payload)
		if err != nil {
			problem(w, http.StatusNotFound, "webhook not found")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"executionId": item.ID, "status": item.Status})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func problem(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"error": detail})
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bytes [16]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			w.Header().Set("X-Request-Id", time.Now().UTC().Format("20060102150405.000000000"))
		} else {
			w.Header().Set("X-Request-Id", hex.EncodeToString(bytes[:]))
		}
		next.ServeHTTP(w, r)
	})
}

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
		next.ServeHTTP(w, r)
	})
}

func cors(next http.Handler) http.Handler {
	allowedOrigin := env("FLOWPILOT_WEB_ORIGIN", "http://localhost:3100")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if origin != allowedOrigin {
				if r.Method == http.MethodOptions {
					problem(w, http.StatusForbidden, "origin not allowed")
					return
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
			}
		}
		if r.Method == http.MethodOptions {
			if origin == "" {
				problem(w, http.StatusForbidden, "origin required")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
