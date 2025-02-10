package proxy

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mittwald/mstudio-ext-proxy/pkg/authentication"
	"github.com/mittwald/mstudio-ext-proxy/pkg/controller"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	Configuration             Configuration
	SessionRepository         repository.SessionRepository
	AuthenticationOptions     authentication.Options
	Logger                    *slog.Logger
	HTTPClient                *http.Client
	RedirectOnUnauthenticated string
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	authCookie, err := request.Cookie(h.AuthenticationOptions.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// TODO
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, session.IssueClaims())
	tokenStr, err := token.SignedString(h.AuthenticationOptions.JWTSecret)
	if err != nil {
		h.responseError(writer, http.StatusInternalServerError, "internal server error", err)
		return
	}

	proxyRequestURL := *request.URL
	proxyRequestURL.Host = h.Configuration.UpstreamURL.Host
	proxyRequestURL.Scheme = h.Configuration.UpstreamURL.Scheme
	proxyRequestURL.User = h.Configuration.UpstreamURL.User

	if h.Configuration.StripPrefix != "" {
		proxyRequestURL.Path = strings.TrimPrefix(proxyRequestURL.Path, h.Configuration.StripPrefix)
	}

	proxyRequest, _ := http.NewRequest(request.Method, proxyRequestURL.String(), request.Body)
	proxyRequest.Header.Set("X-Mstudio-User", tokenStr)
	copyHeaders(request.Header, proxyRequest.Header)

	proxyResponse, err := h.HTTPClient.Do(proxyRequest)
	if err != nil {
		h.responseError(writer, http.StatusBadGateway, "bad gateway", err)
		return
	}

	copyHeaders(proxyResponse.Header, writer.Header())

	writer.WriteHeader(proxyResponse.StatusCode)

	if _, err := io.Copy(writer, proxyResponse.Body); err != nil {
		return
	}
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
