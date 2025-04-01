package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/gofermart/config"
	"github.com/morzisorn/gofermart/internal/client"
	"github.com/morzisorn/gofermart/internal/controllers"
	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/repositories"
	"github.com/morzisorn/gofermart/internal/services/orders"
	"github.com/morzisorn/gofermart/internal/services/processing"
	"github.com/morzisorn/gofermart/internal/services/users"
	"go.uber.org/zap"
)

var cnfg *config.Config

func main() {
	cnfg = config.GetConfig()

	repo := repositories.NewRepository(cnfg)

	userService := users.NewUserService(repo)
	userController := controllers.NewUserController(userService)

	orderService := orders.NewOrderService(repo, userService)
	orderController := controllers.NewOrderController(orderService)

	client := client.NewClient(cnfg)

	processingService := processing.NewProcessingService(orderService, client)

	mux := createServer(userController, orderController)

	if err := runAccrualServer(); err != nil {
		logger.Log.Panic("Error running accrual server", zap.Error(err))
	}

	go runProcessing(processingService, cnfg)

	if err := runServer(mux); err != nil {
		logger.Log.Panic("Error running server", zap.Error(err))
	}
}

func runAccrualServer() error {
	root, err := config.GetProjectRoot()
	if err != nil {
		return fmt.Errorf("get project root error: %w", err)
	}

	filename := fmt.Sprintf("accrual_%s_%s", runtime.GOOS, runtime.GOARCH)

	filepath := filepath.Join(root, "cmd", "accrual", filename)

	cmd := exec.Command(filepath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

func createServer(
	uc *controllers.UserController,
	oc *controllers.OrderController,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()

	mux.POST("/api/user/register", uc.RegisterUser)
	mux.POST("/api/user/login", uc.Login)

	authGroup := mux.Group("/api/user", controllers.AuthMiddleware())
	{
		authGroup.GET("/balance", uc.GetBalance)

		authGroup.POST("/orders", oc.UploadOrder)
		authGroup.POST("/balance/withdraw", oc.Withdraw)
		authGroup.GET("/orders", oc.GetUserOrders)
		authGroup.GET("/withdrawals", oc.GetUserWithdrawals)
	}

	return mux
}

func runServer(mux *gin.Engine) error {
	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Info("Starting server on ", zap.String("address", cnfg.RunAddress))

	return mux.Run(cnfg.RunAddress)
}

func runProcessing(ps *processing.ProcessingService, cnfg *config.Config) {
	lastUpdate := time.Now()
	for {
		if time.Since(lastUpdate).Seconds() > float64(cnfg.LoyaltyUpdateInterval) {
			lastUpdate = time.Now()
			err := ps.ProcessOrders(context.Background())
			if err != nil {
				logger.Log.Error("Processing error: ", zap.Error(err))
			}
		}
	}
}
