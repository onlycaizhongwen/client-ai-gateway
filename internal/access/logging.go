package access

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type logContextKey struct{}

type requestLogInfo struct {
	traceID string
	appID   string
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func (h *Handler) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		info := &requestLogInfo{}
		ctx := context.WithValue(r.Context(), logContextKey{}, info)
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r.WithContext(ctx))
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		h.logger.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Int("bytes", recorder.bytes),
			slog.Duration("duration", time.Since(start)),
			slog.String("trace_id", info.traceID),
			slog.String("app_id", info.appID),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}

func getLogInfo(r *http.Request) *requestLogInfo {
	info, ok := r.Context().Value(logContextKey{}).(*requestLogInfo)
	if !ok || info == nil {
		return &requestLogInfo{}
	}
	return info
}
