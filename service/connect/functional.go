package connect

import (
	"context"
	"log"
	"math/rand"

	"github.com/fooage/messier/event"
	"github.com/fooage/messier/proto/pb"
	"google.golang.org/grpc"
)

// ConnectCluster function is mainly responsible for connecting to the cluster,
// the connection mechanism has been described in the protobuf.
func (r *Router) ConnectCluster(addr string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	r.backup = append(r.backup, addr)
	r.event <- struct{}{}
	for {
		// Use the empty channel feature to block the flow until the next node's disconection.
		<-r.event
		choice := rand.Intn(len(r.backup))
		addr = r.backup[choice]
		conn, err := grpc.Dial(addr, opts...)
		if err != nil {
			log.Fatalf("can't dial to %s: %v\n", addr, err)
		}
		client := pb.NewRouterClient(conn)
		reply, err := client.MoveGuide(context.Background(), &pb.InfoRequest{Info: r.info})
		if err != nil {
			log.Fatalf("rpc call function occur an error: %v\n", err)
		}
		// Successfully find the entry point and nobtify the other service.
		r.info.NextAddress = reply.GetInfo().NextAddress
		r.info.NextHash = reply.GetInfo().NextHash
		r.info.PrevAddress = reply.GetInfo().CurrAddress
		r.info.PrevHash = reply.GetInfo().CurrHash
		event.Publish("change", struct{}{})
		conn, err = grpc.Dial(reply.GetInfo().CurrAddress, opts...)
		if err != nil {
			log.Fatalf("can't dial to %s: %v\n", addr, err)
		}
		client = pb.NewRouterClient(conn)
		_, err = client.SetConnection(context.Background(), &pb.InfoRequest{Info: &pb.Information{
			NextAddress: r.info.CurrAddress,
			NextHash:    r.info.CurrHash,
			PrevAddress: reply.GetInfo().PrevAddress,
			PrevHash:    reply.GetInfo().PrevHash,
		}})
		if err != nil {
			log.Fatalf("rpc set connection error: %v\n", err)
		}
	}
}
