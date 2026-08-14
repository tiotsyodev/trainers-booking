package main

import (
	trainerv1 "booker/gen/trainer/v1"
	"booker/trainer-service/internal/repository/postgres"
	grpcTransport "booker/trainer-service/internal/transport/grpc"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment")
	}

	repoConfig, grpcConfig := postgres.ConfigMustLoad(), grpcTransport.ConfigMustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewConnPool(ctx, repoConfig)
	if err != nil {
		log.Fatalf("error starting connection pool: %v", err)
	}
	defer pool.Close()

	handler := grpcTransport.NewTrainerHandler(&pool, &pool)

	server := grpc.NewServer(grpc.UnaryInterceptor(recovery.UnaryServerInterceptor()))

	trainerv1.RegisterTrainerServiceServer(server, handler)

	lis, err := net.Listen("tcp", ":"+grpcConfig.Port)
	if err != nil {
		log.Fatalf("error starting listener: %v", err)
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()

	server.GracefulStop()

}
