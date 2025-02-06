package repository

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
)

type ExtensionInstanceRepository interface {
	FindExtensionInstanceByID(context.Context, string) (model.ExtensionInstance, error)
	AddExtensionInstance(context.Context, model.ExtensionInstance) error
	UpdateExtensionInstance(context.Context, model.ExtensionInstance) error
	RemoveExtensionInstance(context.Context, model.ExtensionInstance) error
	RemoveExtensionInstanceByID(context.Context, string) error
}
