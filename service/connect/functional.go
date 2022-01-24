package connect

import (
	"context"
	"log"

	"github.com/fooage/messier/event"
	"github.com/fooage/messier/proto/pb"
	"google.golang.org/grpc"
)

// ConnectCluster function is mainly responsible for connecting to the cluster,
// the connection mechanism has been described in the protobuf.
func (r *Router) ConnectCluster(addr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("can't dial to %s: %v\n", addr, err)
	}
	client := pb.NewRouterClient(conn)
	reply, err := client.MoveGuide(context.Background(), &pb.InfoRequest{Info: r.info})
	if err != nil {
		log.Fatalf("rpc call function occur an error: %v\n", err)
	}
	// Successfully find the entry point and nobtify the other service and
	// set the backup information.
	r.info.NextAddress = reply.GetInfo().NextAddress
	r.info.NextHash = reply.GetInfo().NextHash
	r.info.PrevAddress = reply.GetInfo().CurrAddress
	r.info.PrevHash = reply.GetInfo().CurrHash
	event.Publish("redirect", struct{}{})
	prevConn, err := grpc.Dial(reply.GetInfo().CurrAddress, opts...)
	if err != nil {
		log.Fatalf("can't dial to %s: %v\n", addr, err)
	}
	defer prevConn.Close()
	client = pb.NewRouterClient(prevConn)
	_, err = client.SetConnection(context.Background(), &pb.InfoRequest{Info: &pb.Information{
		NextAddress: r.info.CurrAddress,
		NextHash:    r.info.CurrHash,
	}})
	if err != nil {
		log.Fatalf("rpc set connection error: %v\n", err)
	}
	nextConn, err := grpc.Dial(reply.GetInfo().NextAddress, opts...)
	if err != nil {
		log.Fatalf("can't dial to %s: %v\n", addr, err)
	}
	defer nextConn.Close()
	client = pb.NewRouterClient(nextConn)
	_, err = client.SetConnection(context.Background(), &pb.InfoRequest{Info: &pb.Information{
		PrevAddress: r.info.CurrAddress,
		PrevHash:    r.info.CurrHash,
	}})
	if err != nil {
		log.Fatalf("rpc set connection error: %v\n", err)
	}
	go r.ReconnectCluster()
}

// ReconnectCluster function will refresh the backup information while
// heartbeating and reconnect the backup address to let the cluster safe.
func (r *Router) ReconnectCluster() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	for {
		select {
		// update the backup information
		case <-r.refresh:
			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			conn, err := grpc.Dial(r.info.NextAddress, opts...)
			if err != nil {
				log.Fatalf("set backup info error: %v\n", err)
			}
			client := pb.NewRouterClient(conn)
			reply, err := client.GetInformation(context.Background(), &pb.EmptyMessage{})
			if err != nil {
				log.Fatalf("getting info from next node error: %v\n", err)
			}
			r.backup = reply.GetInfo()
		// waitting for the reconnect event
		case <-r.reconnect:
			r.info.NextAddress = r.backup.NextAddress
			r.info.NextHash = r.backup.NextHash
			event.Publish("redirect", struct{}{})
			log.Printf("reconnect to %s\n", r.info.NextAddress)
			conn, err := grpc.Dial(r.info.NextAddress, opts...)
			if err != nil {
				log.Fatalf("reconnect occured an error: %v\n", err)
			}
			client := pb.NewRouterClient(conn)
			client.SetConnection(context.Background(), &pb.InfoRequest{
				Info: &pb.Information{
					PrevAddress: r.info.CurrAddress,
					PrevHash:    r.info.CurrHash,
				},
			})
		}
	}
}
