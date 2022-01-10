package main

import (
	mb "github.com/fluidshare/api-infra/mainbuilder"
	btv1 "github.com/fluidshare/service-clients-internal/devops-domain/backstage-template/v1"
	c "github.com/fluidtruck/deepcopy/controller"
	h "github.com/fluidtruck/deepcopy/handler"
)

func main() {
	var (
		backstageTemplateController c.BackstageTemplateController
	)

	b := mb.NewMainBuilder(&mb.MainBuilderConfig{
		ApplicationName: "deepcopy",
		ControllerLayerConfig: &mb.ControllerLayerConfig{
			CreateControllerLayer: func(b mb.MainBuilder) {
				backstageTemplateController = c.NewBackstageTemplateController(b.GetLogger())
			},
		},
		HandlerLayerConfig: &mb.HandlerLayerConfig{
			CreateRpcHandlers: func(b mb.MainBuilder) {
				btv1.RegisterBackstageTemplateServer(b.GetRpcServer(), h.NewBackstageTemplateHandler(b.GetLogger(), backstageTemplateController))
			},
		},
	})

	defer b.Close()

	b.Run()
}
