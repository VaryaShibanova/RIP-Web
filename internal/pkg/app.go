package pkg

import (
	"fmt"

	"RIP-WEB/internal/app/config"
	"RIP-WEB/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	// Добавляем функции для шаблонов
	a.Router.SetFuncMap(map[string]interface{}{
		"add": func(a, b int) int {
			return a + b
		},
	})

	// Регистрируем API handlers (REST endpoints)
	a.Handler.RegisterAPIHandlers(a.Router)

	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)

	logrus.Infof("Starting server on %s", serverAddress)
	logrus.Info("API endpoints available at:")
	logrus.Info("  GET  /api/anomalies - список аномалий")
	logrus.Info("  POST /api/anomalies - создание аномалии")
	logrus.Info("  GET  /api/trees - список заявок")
	logrus.Info("  GET  /api/trees/cart - корзина")
	logrus.Info("  POST /api/users/register - регистрация")
	logrus.Info("  ... и другие методы")

	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
