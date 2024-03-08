package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/emorydu/grpc-examples/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LaptopClient is a client to call laptop service RPCs
type LaptopClient struct {
	service pb.LaptopServiceClient
}

// NewLaptopClient returns a  new laptop client
func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(cc)
	return &LaptopClient{service: service}
}

// CreateLaptop calls create laptop RPC
func (laptopClient *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	//laptop.Id = ""
	req := &pb.CreateLaptopRequest{Laptop: laptop}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}

// SearchLaptop calls search laptop RPC
func (laptopClient *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Println("search filter: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.service.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Println("- found: ", laptop.GetId())
		log.Println("  + brand: ", laptop.GetBrand())
		log.Println("  + name: ", laptop.GetName())
		log.Println("  + cpu cores: ", laptop.GetCpu())
		log.Println("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Println("  + ram: ", laptop.GetRam())
		log.Println("  + price: ", laptop.GetPriceUsd(), " usd")
	}
}

// UploadImage calls upload image RPC
func (laptopClient *LaptopClient) UploadImage(laptopID string, imagePath string) {
	f, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file:", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image:", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info:", err)
	}

	reader := bufio.NewReader(f)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer:", err, stream.RecvMsg(nil))
		}

		req := &pb.UploadImageRequest{Data: &pb.UploadImageRequest_ChunkData{ChunkData: buffer[:n]}}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server:", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response:", err)
	}

	log.Printf("iamge uploaded with id: %s, size: %d\n", res.GetId(), res.GetSize())
}

// RateLaptop calls rate laptop RPC
func (laptopClient *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Println("no more responses")
				waitResponse <- nil
				return
			}

			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %w", err)
				return
			}

			log.Println("received response:", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}
		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Println("sent request:", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitResponse
	return err
}
