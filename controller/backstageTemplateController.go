package controller

import (
	"context"

	"github.com/sirupsen/logrus"
)

type BackstageTemplateController interface {
	HelloWorld(ctx context.Context) (error)
}

func NewBackstageTemplateController(logger *logrus.Entry) BackstageTemplateController {
	return &backstageTemplateController{
		logger: logger.WithField("struct", "backstageTemplateController"),
	}
}

type backstageTemplateController struct {
	logger *logrus.Entry
}

func (c *backstageTemplateController) HelloWorld(ctx context.Context) (err error) {
	c.logger.Debug("HelloWorld")
	return nil
}
