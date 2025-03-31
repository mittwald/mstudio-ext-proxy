package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/userclientv2"
	"github.com/mittwald/mstudio-ext-proxy/pkg/authentication"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/service"
)

type UserAuthenticationController struct {
	Client                generatedv2.Client
	SessionRepository     repository.SessionRepository
	SessionService        service.SessionService
	InstanceRepository    repository.ExtensionInstanceRepository
	Development           bool
	AuthenticationOptions authentication.Options
	Logger                *slog.Logger
}

type PasswordFormInput struct {
	Password string `form:"password"`
}

func (c *UserAuthenticationController) HandleAuthenticationRequest(ctx *gin.Context) {
	l := c.Logger

	userID, instanceID, atrek, err := extractAuthenticationParamsFromRequest(ctx.Request)
	if err != nil {
		l.Error("failed to extract authentication parameters", "error", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponseFromErr("could not retrieve instance", err))
		return
	}

	l = l.With("userID", userID, "instanceID", instanceID)

	session, err := c.SessionService.InitializeSessionFromRetrievalKey(ctx, atrek, userID, instanceID)
	if err != nil {
		l.Error("failed to create session", "error", err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponseFromErr("error initializing session", err))
		return
	}

	if c.Development {
		ctx.SetCookie(c.AuthenticationOptions.CookieName, session.CookieString(), 0, "/", "", false, false)
	} else {
		ctx.SetCookie(c.AuthenticationOptions.CookieName, session.CookieString(), 0, "/", "", true, true)
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (c *UserAuthenticationController) HandlePasswordAuthentication(ctx *gin.Context) {
	if ctx.Request.Method == http.MethodPost {
		input := PasswordFormInput{}
		if err := ctx.Bind(&input); err != nil {
			ctx.String(http.StatusBadRequest, "missing 'password' parameter")
			return
		}

		if input.Password == c.AuthenticationOptions.StaticPassword {
			session, err := c.buildFakeSession()
			if err != nil {
				ctx.String(http.StatusInternalServerError, "error initializing session")
				return
			}

			if err := c.SessionRepository.CreateSessionWithUnhashedSecret(ctx, session); err != nil {
				ctx.JSON(http.StatusInternalServerError, ErrorResponseFromErr("error initializing session", err))
				return
			}

			ctx.SetCookie(c.AuthenticationOptions.CookieName, session.CookieString(), 3600, "/", "", false, false)
			ctx.Redirect(http.StatusSeeOther, "/")
			return
		}
	}

	ctx.HTML(http.StatusOK, "login.html", gin.H{
		"LoginRoute": "/mstudio/auth/password",
	})
}

// CAUTION: DO NOT USE IN PRODUCTION
func (c *UserAuthenticationController) HandleFakeAuthentication(ctx *gin.Context) {
	if !c.Development {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "not available"})
		return
	}

	session, err := c.buildFakeSession()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponseFromErr("error initializing session", err))
		return
	}

	if err := c.SessionRepository.CreateSessionWithUnhashedSecret(ctx, session); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponseFromErr("error initializing session", err))
		return
	}

	ctx.SetCookie(c.AuthenticationOptions.CookieName, session.CookieString(), 3600, "/", "", false, false)
	ctx.Redirect(http.StatusSeeOther, "/")
}

func (c *UserAuthenticationController) getAPITokenFromATREK(ctx context.Context, atrek, userID string) (string, string, time.Time, error) {
	req := userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequest{
		Body: userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequestBody{
			AccessTokenRetrievalKey: atrek,
			UserId:                  userID,
		},
	}

	resp, _, err := c.Client.User().AuthenticateWithAccessTokenRetrievalKey(ctx, req)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return resp.Token, resp.RefreshToken, resp.ExpiresAt, nil
}

func (c *UserAuthenticationController) buildFakeSession() (model.Session, error) {
	session, err := model.NewSession()
	if err != nil {
		return session, err
	}

	session.Expires = time.Now().Add(c.AuthenticationOptions.CookieTTL)
	session.Email = "user@mstudio.example"
	session.UserID = "522963df-3ebf-4158-80cc-1e9a78aca9b5"
	session.FirstName = "Max"
	session.LastName = "Mustermann"
	session.AccessToken = "fake-api-token"
	session.Instance.ID = "848821a6-7bbb-4b15-a267-7b67e14e5a27"
	session.Instance.Enabled = true
	session.Instance.Context.Kind = "customer"
	session.Instance.Context.ID = "4a30329f-3bb7-4871-b9e2-e4815718e74a"
	session.Instance.Scopes = []string{}
	session.Instance.Secret = []byte("very secret")

	return session, nil
}

// extractAuthenticationParamsFromRequest is a helper function to extract all
// necessary authentication options from a request. This is done by directly
// iterating over the query params in order to allow case-insensitive params
// (to avoid common confusions like "userId" vs "userID")
func extractAuthenticationParamsFromRequest(req *http.Request) (userID, instanceID, atrek string, err error) {
	for key, values := range req.URL.Query() {
		switch strings.ToLower(key) {
		case "userid":
			userID = values[0]
		case "instanceid":
			instanceID = values[0]
		case "atrek":
			atrek = values[0]
		}
	}

	if userID == "" || instanceID == "" || atrek == "" {
		err = fmt.Errorf("all the userId, instanceId and atrek query params must be set")
	}

	return
}
