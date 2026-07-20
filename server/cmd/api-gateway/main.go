package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-core-banking-system/internal/api-gateway/client"
	"go-core-banking-system/internal/api-gateway/handler"
	"go-core-banking-system/internal/api-gateway/router"
	"go-core-banking-system/internal/api-gateway/validator"
	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

func main() {
	accountAddr := getEnv("ACCOUNT_SERVICE_ADDR", "localhost:50051")
	transactionAddr := getEnv("TRANSACTION_SERVICE_ADDR", "localhost:50052")
	httpPort := getEnv("HTTP_PORT", "8080")

	accountConn, err := grpc.NewClient(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to account service: %v", err)
	}
	defer func() { _ = accountConn.Close() }()

	transactionConn, err := grpc.NewClient(transactionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to transaction service: %v", err)
	}
	defer func() { _ = transactionConn.Close() }()

	accountClient := client.NewAccountClientWrapper(pb_account.NewAccountServiceClient(accountConn))
	transactionClient := client.NewTransactionClientWrapper(pb_transaction.NewTransactionServiceClient(transactionConn))
	v := validator.New()

	h := handler.NewHandler(accountClient, transactionClient, v)
	r := router.NewRouter(h)

	addr := net.JoinHostPort("", httpPort)
	fmt.Printf("API Gateway listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := getEnvRaw(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvRaw(key string) string {
	return ""
}
