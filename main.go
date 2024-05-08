package main

import (
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/shutdown"
	"isp-config-service/conf"
	"isp-config-service/routes"
	"isp-config-service/service/startup"
)

var (
	version = "1.0.0"
)

// @title isp-config-service
// @version 1.0.0
// @description Модуль управления конфигурациями

// @license.name GNU GPL v3.0

// @host localhost:9000
// @BasePath /api/config

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json docs/docs.go
func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors())
	app := boot.App
	logger := app.Logger()

	startup := startup.New(boot)
	app.AddRunners(startup)
	app.AddClosers(startup.Closers()...)

	shutdown.On(func() {
		logger.Info(app.Context(), "starting shutdown")
		app.Shutdown()
		logger.Info(app.Context(), "shutdown completed")
	})

	err := app.Run()
	if err != nil {
		boot.Fatal(err)
	}
}
