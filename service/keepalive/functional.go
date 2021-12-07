package keepalive

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/fooage/messier/event"
	"github.com/fooage/messier/proto/pb"
	"google.golang.org/grpc"
)

// HeartKeeping function ensures that it always listens to the next node's
// keepalive packet and sends a keepalive to the new node if the connection
// information has been updated.
func (h Heartbeat) HeartKeeping() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	for {
		conn, err := grpc.Dial(h.info.NextAddress, opts...)
		if err != nil {
			log.Fatalf("failed to dial: %v\n", err)
		}
		client := pb.NewHeartbeatClient(conn)
		stream, err := client.HeartWatch(context.Background(), &pb.HeartbeatRequest{Code: 200})
		if err != nil {
			log.Fatalf("function call error: %v\n", err)
		}
		var lastTime = time.Now()
	receiving:
		for {
			select {
			case <-h.event:
				break receiving
			default:
				reply, err := stream.Recv()
				// Disconnect from the next node if any of the connection problems occur.
				if err == io.EOF || err != nil || time.Now().After(lastTime.Add(h.timeout)) {
					log.Printf("the %s occured an timeout: %v\n", h.info.NextAddress, err)
					event.Publish("reconnect", struct{}{})
					break receiving
				}
				lastTime = time.Now()
				log.Printf("recv a heartbeat from %s: %v\n", h.info.NextAddress, reply)
			}
		}
	}
}
