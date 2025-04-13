package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
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

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}
	cnfg := config.GetConfig()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	var accrualCmd *exec.Cmd

	repo := repositories.NewRepository(cnfg)

	userService := users.NewUserService(repo)
	userController := controllers.NewUserController(userService)

	orderService := orders.NewOrderService(repo, userService)
	orderController := controllers.NewOrderController(orderService)

	client := client.NewClient(cnfg)

	processingService := processing.NewProcessingService(orderService, client)

	mux := createServer(userController, orderController)

	accrualCmd, err := createAccrualServer(cnfg)
	//defer killProcess(accrualCmd)

	if err != nil {
		logger.Log.Fatal("Failed to create accrual server")
	}

	if err := accrualCmd.Start(); err != nil {
		logger.Log.Fatal("Failed to start accrual server. ", zap.Error(err))
	}

	go runProcessing(context.Background(), processingService, cnfg)

	if err := runServer(mux, cnfg); err != nil {
		logger.Log.Error("Error running server", zap.Error(err))
		err := accrualCmd.Process.Kill()
		if err != nil {
			logger.Log.Error("Failed to kill accrual process", zap.Error(err))
		} else {
			logger.Log.Info("Accrual process killed")
		}
	}
}

func createAccrualServer(cnfg *config.Config) (*exec.Cmd, error) {
	root, err := config.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("get project root error: %w", err)
	}

	filename := fmt.Sprintf("accrual_%s_%s", runtime.GOOS, runtime.GOARCH)
	path := filepath.Join(root, "cmd", "accrual", filename)
	logger.Log.Info("Accrual binary path", zap.String("path", path))

	accrualCmd := exec.Command(path, "-a", cnfg.AccrualSystemAddress)
	accrualCmd.Env = []string{}

	accrualCmd.Stdout = os.Stdout
	accrualCmd.Stderr = os.Stderr

	return accrualCmd, nil
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

		authGroup.POST("/orders", controllers.RequireContentType("text/plain"), oc.UploadOrder)
		authGroup.POST("/balance/withdraw", oc.Withdraw)
		authGroup.GET("/orders", oc.GetUserOrders)
		authGroup.GET("/withdrawals", oc.GetUserWithdrawals)
	}

	return mux
}

func runServer(mux *gin.Engine, cnfg *config.Config) error {
	logger.Log.Info("Starting server on ", zap.String("address", cnfg.RunAddress))

	return mux.Run(cnfg.RunAddress)
}

func runProcessing(ctx context.Context, ps *processing.ProcessingService, cnfg *config.Config) {
	ticker := time.NewTicker(time.Duration(cnfg.LoyaltyUpdateInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Context canceled, stopping processing loop")
			return
		case <-ticker.C:
			err := ps.ProcessOrders(context.Background())
			if err != nil {
				logger.Log.Error("Processing error: ", zap.Error(err))
			}
		}
	}
}


func killProcess(p *exec.Cmd) {
	if p != nil && p.Process != nil {
		err := p.Process.Kill()
		if err != nil {
			logger.Log.Error("Failed to kill accrual process", zap.Error(err))
		} else {
			logger.Log.Info("Accrual process killed")
		}
	}
}