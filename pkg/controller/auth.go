package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mittwald/api-client-go/mittwaldv2"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/userclientv2"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"net/http"
	"time"
)

type UserAuthenticationController struct {
	Client            generatedv2.Client
	SessionRepository repository.SessionRepository
	Development       bool
	TTL               time.Duration
}

func (c *UserAuthenticationController) HandleAuthenticationRequest(ctx *gin.Context) {
	atrek, ok := ctx.GetQuery("atrek")
	if !ok {
		ctx.JSON(http.StatusBadRequest, errorResponse{Message: "missing 'atrek' query parameter"})
		return
	}

	userID, ok := ctx.GetQuery("userID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, errorResponse{Message: "missing 'userID' query parameter"})
		return
	}

	token, err := c.getAPITokenFromATREK(ctx, atrek, userID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponseFromErr("error getting access token", err))
		return
	}

	authClient, err := mittwaldv2.New(ctx, mittwaldv2.WithAccessToken(token))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponseFromErr("error authenticating at API", err))
		return
	}

	req := userclientv2.GetUserRequest{UserID: userID}
	resp, _, err := authClient.User().GetUser(ctx, req)
	if err != nil {
		return
	}

	session := model.Session{
		ID:          uuid.Must(uuid.NewUUID()).String(),
		Expires:     time.Now().Add(c.TTL),
		Email:       strPtrOr(resp.Email, ""),
		UserID:      resp.UserId,
		FirstName:   resp.Person.FirstName,
		LastName:    resp.Person.LastName,
		AccessToken: token,
	}

	if err := c.SessionRepository.CreateSessionWithUnhashedSecret(ctx, session); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponseFromErr("error initializing session", err))
		return
	}

	if c.Development {
		ctx.SetCookie("mstudio_ext_session", session.CookieString(), 3600, "/", "", false, false)
	} else {
		ctx.SetCookie("mstudio_ext_session", session.CookieString(), 3600, "/", "", true, true)
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (c *UserAuthenticationController) getAPITokenFromATREK(ctx context.Context, atrek, userID string) (string, error) {
	req := userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequest{
		Body: userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequestBody{
			AccessTokenRetrievalKey: atrek,
			UserId:                  userID,
		},
	}

	resp, _, err := c.Client.User().AuthenticateWithAccessTokenRetrievalKey(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func strPtrOr(one *string, alt string) string {
	if one != nil {
		return *one
	}
	return alt
}
