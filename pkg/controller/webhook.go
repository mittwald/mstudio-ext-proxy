package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks"
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhookscommon"
	"github.com/mittwald/mstudio-ext-proxy/pkg/webhooks/webhooksv1"
	"io"
	"log/slog"
	"net/http"
)

type WebhookController struct {
	ExtensionInstanceRepository repository.ExtensionInstanceRepository
	WebhookVerifier             *webhookscommon.Verifier
	Logger                      *slog.Logger
}

func (c *WebhookController) HandleWebhookRequest(ctx *gin.Context) {
	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponseFromErr("error reading payload", err))
		return
	}

	if err := c.WebhookVerifier.VerifyWebhookRequest(ctx, ctx.Request, payload); err != nil {
		c.Logger.Debug("invalid webhook signature", "err", err)
		ctx.JSON(http.StatusForbidden, errorResponseFromErr("invalid request signature", err))
		return
	}

	wh, env, err := webhooks.UnmarshalWebhookRequest(payload)
	if err != nil {
		c.Logger.Debug("invalid webhook request", "err", err)
		ctx.JSON(http.StatusBadRequest, errorResponseFromErr("could not decode webhook request", err))
		return
	}

	c.Logger.Debug("handling webhook", "webhook.kind", env.Kind, "webhook.version", env.APIVersion)

	switch wht := wh.(type) {
	case *webhooksv1.ExtensionAddedToContext:
		err = c.handleExtensionAddedToContextV1(ctx, wht)
	case *webhooksv1.ExtensionInstanceUpdated:
		err = c.handleExtensionUpdatedV1(ctx, wht)
	case *webhooksv1.ExtensionInstanceSecretRotated:
		err = c.handleExtensionSecretRotatedV1(ctx, wht)
	case *webhooksv1.ExtensionInstanceRemovedFromContext:
		err = c.handleExtensionInstanceFromvedFromContextV1(ctx, wht)
	}

	if err != nil {
		c.Logger.Debug("invalid webhook request", "err", err)
		ctx.JSON(http.StatusBadRequest, errorResponseFromErr("could not decode webhook request", err))
		return
	}

	ctx.JSON(http.StatusOK, payload)
}

func (c *WebhookController) handleExtensionAddedToContextV1(ctx context.Context, wh *webhooksv1.ExtensionAddedToContext) error {
	instance := model.ExtensionInstance{
		ID: wh.ID,
		Context: model.ExtensionInstanceContext{
			ID:   wh.Context.ID,
			Kind: string(wh.Context.Kind),
		},
		Enabled: wh.State.Enabled,
		Scopes:  wh.ConsentedScopes,
		Secret:  []byte(wh.Secret),
	}

	return c.ExtensionInstanceRepository.AddExtensionInstance(ctx, instance)
}

func (c *WebhookController) handleExtensionUpdatedV1(ctx context.Context, wh *webhooksv1.ExtensionInstanceUpdated) error {
	instance, err := c.ExtensionInstanceRepository.FindExtensionInstanceByID(ctx, wh.ID)
	if err != nil {
		return err
	}

	instance.Scopes = wh.ConsentedScopes
	instance.Enabled = wh.State.Enabled

	return c.ExtensionInstanceRepository.UpdateExtensionInstance(ctx, instance)
}

func (c *WebhookController) handleExtensionSecretRotatedV1(ctx context.Context, wh *webhooksv1.ExtensionInstanceSecretRotated) error {
	instance, err := c.ExtensionInstanceRepository.FindExtensionInstanceByID(ctx, wh.ID)
	if err != nil {
		return err
	}

	instance.Secret = []byte(wh.Secret)

	return c.ExtensionInstanceRepository.UpdateExtensionInstance(ctx, instance)
}

func (c *WebhookController) handleExtensionInstanceFromvedFromContextV1(ctx context.Context, wh *webhooksv1.ExtensionInstanceRemovedFromContext) error {
	return c.ExtensionInstanceRepository.RemoveExtensionInstanceByID(ctx, wh.ID)
}
