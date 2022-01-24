package keepalive

import (
	"context"
	"log"
	"time"

	"github.com/fooage/messier/event"
	"github.com/fooage/messier/proto/pb"
	"github.com/spf13/viper"
)

// The Heartbeat service will act as a subscriber to listen for changes in the
// router's connection. Timeout set in config file.
type Heartbeat struct {
	info    *pb.Information
	timeout time.Duration
	// This interface is very important and deals with listening for changes
	// in connection information within this node so that other services can
	// handle it correctly.
	redirect event.EventChannel
	pb.UnimplementedHeartbeatServer
}

func NewHeartbeat(info *pb.Information) *Heartbeat {
	redirectChannel := make(event.EventChannel, 8)
	event.Subscribe("redirect", redirectChannel)
	return &Heartbeat{info: info,
		timeout:  time.Second * time.Duration(viper.GetInt64("server.keepalive.time_out")),
		redirect: redirectChannel,
	}
}

// HeartCheck is a single check function which will send a keepalive.
func (h Heartbeat) HeartCheck(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	code := request.GetCode()
	if code != 200 {
		return &pb.HeartbeatReply{Status: pb.HeartbeatReply_UNKNOWN}, nil
	} else {
		return &pb.HeartbeatReply{Status: pb.HeartbeatReply_SERVING}, nil
	}
}

// HeartWatch function will always maintain a stream to send heartbeats.
func (h Heartbeat) HeartWatch(request *pb.HeartbeatRequest, server pb.Heartbeat_HeartWatchServer) error {
	trick := time.NewTicker(h.timeout / 2)
	// There's time be set in half of time out duration.
	defer trick.Stop()
	for {
		err := server.Send(&pb.HeartbeatReply{Status: pb.HeartbeatReply_SERVING})
		if err != nil {
			log.Printf("keepalive has been killed: %v\n", err)
			return err
		}
		<-trick.C
	}
}
