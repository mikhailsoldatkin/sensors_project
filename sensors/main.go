package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "sensors/proto"
)

var (
	password           = os.Getenv("SENSORS_PASSWORD")
	collectorPort      = os.Getenv("COLLECTOR_PORT")
	collectorContainer = os.Getenv("COLLECTOR_CONTAINER")
	numCPU             = runtime.NumCPU()
)

const (
	numberOfSensors    = 5
	dataGenerationRate = 3 * time.Second
	minTemperature     = 10.0
	maxTemperature     = 30.0
)

type Sensor struct {
	ID   int64
	Type int64
}

// generateSensorData emulates working sensors
func generateSensorData(ctx context.Context, sensor Sensor, stream pb.SensorService_StreamSensorDataClient) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			temperature := minTemperature + rand.Float64()*(maxTemperature-minTemperature)
			data := &pb.SensorData{
				SensorId:  sensor.ID,
				Type:      sensor.Type,
				Value:     float32(temperature),
				Timestamp: timestamppb.Now(),
			}
			if err := stream.Send(data); err != nil {
				log.Fatalf("Failed to send data: %v", err)
			}
			time.Sleep(dataGenerationRate)
		}
	}
}

// createSensors creates slice of Sensors
func createSensors(num int) []Sensor {
	sensors := make([]Sensor, 0, num)
	for i := 0; i < num; i++ {
		sensors = append(sensors, Sensor{ID: int64(i), Type: 1})
	}
	return sensors
}

func main() {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	target := fmt.Sprintf("%s:%s", collectorContainer, collectorPort)

	conn, err := grpc.NewClient(target, creds)
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing gRPC connection: %v", err)
		}
	}(conn)

	client := pb.NewSensorServiceClient(conn)

	md := metadata.New(map[string]string{"password": password})
	streamCtx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := client.StreamSensorData(streamCtx)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sensors := createSensors(numberOfSensors)
	jobs := make(chan Sensor)
	wg := sync.WaitGroup{}

	for i := 0; i < numCPU && i < numberOfSensors; i++ {
		go func() {
			defer wg.Done()
			for sensor := range jobs {
				generateSensorData(ctx, sensor, stream)
			}
		}()
	}

	for _, sensor := range sensors {
		wg.Add(1)
		jobs <- sensor
	}

	close(jobs)
	wg.Wait()
}
