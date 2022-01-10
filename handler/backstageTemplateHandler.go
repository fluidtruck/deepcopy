package handler

import (
	"context"

	"github.com/sirupsen/logrus"

	v1 "github.com/fluidshare/service-clients-internal/devops-domain/backstage-template/v1"
	c "github.com/fluidtruck/deepcopy/controller"
)

func NewBackstageTemplateHandler(logger *logrus.Entry, backstageTemplateController c.BackstageTemplateController) v1.BackstageTemplateServer {
	return &backstageTemplateHandler{
		logger:                      logger.WithField("struct", "helloWorldHandler"),
		backstageTemplateController: backstageTemplateController,
	}
}

type backstageTemplateHandler struct {
	v1.UnimplementedBackstageTemplateServer
	logger                  *logrus.Entry
	backstageTemplateController c.BackstageTemplateController
}

func (h *backstageTemplateHandler) HelloWorld(ctx context.Context, request *v1.GetHelloWorldRequest) (*v1.GetHelloWorldResponse, error) {
	logger := h.logger.WithField("request", request)
	logger.Info("HelloWorld")

	err := h.backstageTemplateController.HelloWorld(ctx)
	if err != nil {
		msg := "failed to say hellow"
		logger.WithError(err).Error(msg)
		return &v1.GetHelloWorldResponse{
			Success:  false,
		}, nil
	}

	return &v1.GetHelloWorldResponse{
		Success: true,
	}, nil
}
