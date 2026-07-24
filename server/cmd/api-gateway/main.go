package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-core-banking-system/internal/api-gateway/client"
	"go-core-banking-system/internal/api-gateway/handler"
	"go-core-banking-system/internal/api-gateway/router"
	"go-core-banking-system/internal/api-gateway/validator"
	"go-core-banking-system/pkg/app"
	"go-core-banking-system/pkg/observability"
	"go-core-banking-system/pkg/resilience"
	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

func main() {
	accountAddr := app.GetEnv("ACCOUNT_SERVICE_ADDR", "localhost:50051")
	transactionAddr := app.GetEnv("TRANSACTION_SERVICE_ADDR", "localhost:50052")
	httpPort := app.GetEnv("HTTP_PORT", "8080")
	otlpEndpoint := app.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := observability.Init("api-gateway", otlpEndpoint)
	if err != nil {
		log.Printf("warning: observability init failed: %v", err)
	}
	if shutdown != nil {
		defer func() {
			_ = shutdown(context.Background())
		}()
	}

	accountCB := resilience.New(resilience.DefaultConfig("account-service"))
	transactionCB := resilience.New(resilience.DefaultConfig("transaction-service"))

	accountConn, err := grpc.NewClient(accountAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			resilience.UnaryClientInterceptor(accountCB, 5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to connect to account service: %v", err)
	}
	defer func() { _ = accountConn.Close() }()

	transactionConn, err := grpc.NewClient(transactionAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			resilience.UnaryClientInterceptor(transactionCB, 10*time.Second),
		),
	)
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
