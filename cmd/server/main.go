package main

import (
	"log"
	"net"

	schedulepb "github.com/je0ng3/remindme-cli/api/proto/schedulepb"
	"github.com/je0ng3/remindme-cli/internal/server"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	s := server.NewSchedulerServer("data/schedules.csv")
	schedulepb.RegisterSchedulerServer(grpcServer, s)

	log.Println("Server is running at :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}