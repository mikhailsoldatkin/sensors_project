package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	pb "collector/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedSensorServiceServer
}

var password = "password" // TODO change

func (s *server) StreamSensorData(stream pb.SensorService_StreamSensorDataServer) error {

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	p := md["password"]
	if p[0] != password {
		log.Println("Authentication failed")
		return status.Errorf(codes.Unauthenticated, "wrong password")
	}

	for {
		sensorData, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by client")
			return nil
		}
		if err != nil {
			log.Printf("Failed to receive data: %v", err)
			return err
		}
		log.Printf("Received data: SensorID: %d, Type: %v, Value: %.2f, Timestamp: %s",
			sensorData.SensorId, sensorData.Type, sensorData.Value, sensorData.Timestamp.AsTime().String())
	}
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("COLLECTOR_PORT")))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSensorServiceServer(s, &server{})

	log.Printf("%s service is listening on port %s...", os.Getenv("COLLECTOR_CONTAINER"), os.Getenv("COLLECTOR_PORT"))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
