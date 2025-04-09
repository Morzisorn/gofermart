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

var cnfg *config.Config

var accrualCmd *exec.Cmd

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}
	cnfg = config.GetConfig()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Log.Info("Shutting down...")

		if accrualCmd != nil && accrualCmd.Process != nil {
			err := accrualCmd.Process.Kill()
			if err != nil {
				logger.Log.Error("Failed to kill accrual process", zap.Error(err))
			} else {
				logger.Log.Info("Accrual process killed")
			}
		}

		os.Exit(0)
	}()

	repo := repositories.NewRepository(cnfg)

	userService := users.NewUserService(repo)
	userController := controllers.NewUserController(userService)

	orderService := orders.NewOrderService(repo, userService)
	orderController := controllers.NewOrderController(orderService)

	client := client.NewClient(cnfg)

	processingService := processing.NewProcessingService(orderService, client)

	mux := createServer(userController, orderController)

	if err := runAccrualServer(); err != nil {
		logger.Log.Error("Error running accrual server", zap.Error(err))
	}

	go runProcessing(processingService, cnfg)

	if err := runServer(mux); err != nil {
		logger.Log.Error("Error running server", zap.Error(err))
		err := accrualCmd.Process.Kill()
		if err != nil {
			logger.Log.Error("Failed to kill accrual process", zap.Error(err))
		} else {
			logger.Log.Info("Accrual process killed")
		}
	}
}

func runAccrualServer() error {
	root, err := config.GetProjectRoot()
	if err != nil {
		return fmt.Errorf("get project root error: %w", err)
	}

	filename := fmt.Sprintf("accrual_%s_%s", runtime.GOOS, runtime.GOARCH)
	path := filepath.Join(root, "cmd", "accrual", filename)

	accrualCmd = exec.Command(path, "-a", cnfg.AccrualSystemAddress)
	accrualCmd.Env = []string{}

	accrualCmd.Stdout = os.Stdout
	accrualCmd.Stderr = os.Stderr

	logger.Log.Info("Accrual binary path", zap.String("path", path))

	logger.Log.Info("Running accrual server")
	if err := accrualCmd.Start(); err != nil {
		return fmt.Errorf("failed to start accrual server: %w", err)
	}

	// go func() {
	// 	err := accrualCmd.Wait()
	// 	if err != nil {
	// 		logger.Log.Error("accrual process exited with error", zap.Error(err))
	// 		accrualCmd.Process.Kill()
	// 	}
	// }()

	// const maxAttempts = 3
	// for i := 0; i < maxAttempts; i++ {
	// 	resp, err := http.Get(fmt.Sprintf("http://%s/api/orders/12345678903", cnfg.AccrualSystemAddress))
	// 	if err == nil && resp.StatusCode < 300 {
	// 		logger.Log.Info("Accrual server is up and running")
	// 		return nil
	// 	}
	// 	time.Sleep(500 * time.Millisecond)
	// }

	return fmt.Errorf("accrual server did not start in time")
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
