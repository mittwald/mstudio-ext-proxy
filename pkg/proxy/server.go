package proxy

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mittwald/mstudio-ext-proxy/pkg/authentication"
	"github.com/mittwald/mstudio-ext-proxy/pkg/controller"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
)

type Handler struct {
	Configuration             Configuration
	SessionRepository         repository.SessionRepository
	AuthenticationOptions     authentication.Options
	Logger                    *slog.Logger
	HTTPClient                *http.Client
	RedirectOnUnauthenticated string
	ProxyBufferSize           int64
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	authCookie, err := request.Cookie(h.AuthenticationOptions.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			h.respondUnauthorized(writer)
			return
		}
	}

	sessionID, sessionSecret := model.SessionIDAndSecretFromCookieString(authCookie.Value)
	session, err := h.SessionRepository.FindSessionByIDAndSecret(request.Context(), sessionID, sessionSecret)
	if err != nil {
		h.Logger.Warn("invalid session", "err", err)
		h.respondUnauthorized(writer)
		return
	}

	token, err := h.buildUserJWT(session)
	if err != nil {
		h.responseError(writer, http.StatusInternalServerError, "internal server error", err)
		return
	}

	proxyRequest := h.buildProxyRequest(request, token)
	proxyResponse, err := h.HTTPClient.Do(proxyRequest)
	if err != nil {
		h.responseError(writer, http.StatusBadGateway, "bad gateway", err)
		return
	}

	h.copyProxyResponse(writer, proxyResponse)
}

func (h *Handler) buildUserJWT(session *model.Session) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, session.IssueClaims())
	tokenStr, err := token.SignedString(h.AuthenticationOptions.JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (h *Handler) buildProxyRequest(request *http.Request, tokenStr string) *http.Request {
	proxyRequestURL := h.buildProxyRequestURL(request)

	l := h.Logger.With("req.url", request.URL.String(), "upstream.url", proxyRequestURL)
	l.Debug("proxying request")

	proxyRequest, _ := http.NewRequest(request.Method, proxyRequestURL, request.Body)
	proxyRequest.Header.Set("X-Mstudio-User", tokenStr)
	copyHeaders(request.Header, proxyRequest.Header)
	return proxyRequest
}

func (h *Handler) buildProxyRequestURL(request *http.Request) string {
	proxyRequestURL := *request.URL
	proxyRequestURL.Host = h.Configuration.UpstreamURL.Host
	proxyRequestURL.Scheme = h.Configuration.UpstreamURL.Scheme
	proxyRequestURL.User = h.Configuration.UpstreamURL.User

	if h.Configuration.StripPrefix != "" {
		proxyRequestURL.Path = strings.TrimPrefix(proxyRequestURL.Path, h.Configuration.StripPrefix)
	}

	return proxyRequestURL.String()
}

func (h *Handler) copyProxyResponse(writer http.ResponseWriter, proxyResponse *http.Response) {
	l := h.Logger.With("res.status", proxyResponse.StatusCode)
	l.Debug("proxy response")

	copyHeaders(proxyResponse.Header, writer.Header())

	writer.WriteHeader(proxyResponse.StatusCode)
	if err := h.copyProxyResponseBody(proxyResponse.Body, writer); err != nil {
		h.Logger.Warn("error while copying proxy response", "err", err)
	}
}

func (h *Handler) copyProxyResponseBody(proxyResponse io.Reader, writer io.Writer) error {
	if h.ProxyBufferSize != 0 {
		buf := make([]byte, h.ProxyBufferSize)
		if _, err := io.CopyBuffer(writer, proxyResponse, buf); err != nil {
			return err
		}
	}

	_, err := io.Copy(writer, proxyResponse)
	return err
}

func (h *Handler) responseError(writer http.ResponseWriter, code int, msg string, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	_ = json.NewEncoder(writer).Encode(controller.ErrorResponseFromErr(msg, err))
}

func (h *Handler) respondUnauthorized(writer http.ResponseWriter) {
	if h.AuthenticationOptions.StaticPassword != "" {
		writer.Header().Set("Location", "/mstudio/auth/password")
		writer.WriteHeader(http.StatusSeeOther)
		return
	}

	if h.RedirectOnUnauthenticated != "" {
		writer.Header().Set("Location", h.RedirectOnUnauthenticated)
		writer.WriteHeader(http.StatusSeeOther)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(writer).Encode(controller.ErrorResponse{Message: "unauthorized"})
}

func copyHeaders(source, target http.Header) {
	for header, values := range source {
		for _, value := range values {
			target.Add(header, value)
		}
	}
}
