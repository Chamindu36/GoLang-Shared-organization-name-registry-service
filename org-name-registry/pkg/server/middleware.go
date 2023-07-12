package server

import (
	"bytes"
	"fmt"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/oidc"
	"github.com/Chamindu36/organization-name-registry-service/pkg/logging"
	"github.com/Chamindu36/organization-name-registry-service/pkg/trace"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type MiddlewareConfig struct {
	LoggingMiddlewareConfig LoggingMiddlewareConfig `yaml:"logging"`
}

type LoggingMiddlewareConfig struct {
	DumpRequestBody     bool   `yaml:"dumpRequestBody"`
	DumpRequestHeaders  bool   `yaml:"dumpRequestHeaders"`
	DumpResponseBody    bool   `yaml:"dumpResponseBody"`
	DumpResponseHeaders bool   `yaml:"dumpResponseHeaders"`
	PublicKey           string `yaml:"publicKey"`
}

type responseRecorder struct {
	w                  http.ResponseWriter
	status             int
	body               bytes.Buffer
	size               int
	writeHeaderInvoked bool
}

type OauthConfig struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SubjectClaim string
}

func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	size, err := r.w.Write(b)
	r.size += size
	return size, err
}

func (r *responseRecorder) WriteHeader(s int) {
	r.writeHeaderInvoked = true
	r.w.WriteHeader(s)
	r.status = s
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Body() string {
	return r.body.String()
}

func (r *responseRecorder) Size() int {
	return r.size
}

func LoggingMiddleware(cfg LoggingMiddlewareConfig, logger *zap.SugaredLogger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// avoid logging health probes
			start := time.Now()

			logFieldReqBody := zap.Skip()
			logFieldReqHeaders := zap.Skip()
			logFieldRespBody := zap.Skip()
			logFieldRespHeaders := zap.Skip()

			recorder := &responseRecorder{w: w, status: http.StatusOK}
			next.ServeHTTP(recorder, req)
			duration := time.Since(start)

			go func() {
				ctx := req.Context()
				statusCode := recorder.status
				level := zap.InfoLevel
				var err error
				if statusCode >= 500 {
					err = fmt.Errorf("%s", recorder.Body())
					level = zap.ErrorLevel
				} else if statusCode >= 400 {
					err = fmt.Errorf("%s", recorder.Body())
				}
				logger.Desugar().Check(level, "Logging middleware").Write(
					zap.String("request-id", trace.RequestIdFromContext(ctx)),
					zap.Int("hops", trace.HopCountFromContext(ctx)),
					zap.String("request-path", trace.RequestPathFromContext(ctx)),
					zap.String("method", req.Method),
					zap.String("url", req.URL.String()),
					zap.String("user-agent", req.UserAgent()),
					zap.Int("status", statusCode),
					zap.Int("response-size", recorder.Size()),
					zap.Error(err),
					logFieldReqBody,
					logFieldReqHeaders,
					logFieldRespBody,
					logFieldRespHeaders,
					zap.Time("start-time", start),
					zap.Duration("duration", duration),
				)
			}()
		})
	}
}

func ResponseHeadersMiddleware(headers map[string]string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func TraceMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(trace.NewFromIncomingContext(req.Context())))
		})
	}
}

//oidc middleware to handle authentication process
func OidcMiddleware(authenticator oidc.Interface) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			headerAndBody := ""
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				logging.NewDefaultLogger().Warnf("Authorization header is not provided in the request")
				http.Error(w, "Authorization header is not provided in the request", http.StatusUnauthorized)
				return
			}
			if len(authHeader) > 0 {
				authHeaderParts := strings.Fields(authHeader)
				if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
					logging.NewDefaultLogger().Warnf("Authorization header format must be Bearer {token}")
					http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
					return
				}
				headerAndBody = authHeaderParts[1]
			} else {
				// Authorization header not present, try to get from a cookie
				logging.NewDefaultLogger().Warnf("Authorization header doe not contain all the required parts")
				http.Error(w, "Authorization header doe not contain all the required parts", http.StatusUnauthorized)
				return
			}

			authenticated, error := authenticator.AuthenticateToken(headerAndBody)
			if error != nil {
				http.Error(w, error.Error(), http.StatusUnauthorized)
				return
			}
			if !authenticated {
				logging.NewDefaultLogger().Warnf("Token validation failed")
				http.Error(w, "Token validation failed", http.StatusUnauthorized)
				return
			}
			handler.ServeHTTP(w, req)
		})
	}
}
