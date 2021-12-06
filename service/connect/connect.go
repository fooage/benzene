package connect

import (
	"context"

	"github.com/fooage/messier/event"
	"github.com/fooage/messier/proto/pb"
	"google.golang.org/grpc"
)

// Router has its information which include ip hash and address.
type Router struct {
	info   *pb.Information
	backup []string           // there are some address for reconnect cluster
	event  event.EventChannel // notice router reconnect to other node
	pb.UnimplementedRouterServer
}

func NewRouter(info *pb.Information) *Router {
	ch := make(event.EventChannel, 8)
	event.Subscribe("reconnect", ch)
	return &Router{
		info:   info,
		backup: make([]string, 0, 16),
		event:  ch,
	}
}

func (r Router) GetInformation(ctx context.Context, empty *pb.EmptyMessage) (*pb.InfoReply, error) {
	return &pb.InfoReply{Info: r.info}, nil
}

func (r *Router) SetConnection(ctx context.Context, request *pb.InfoRequest) (*pb.EmptyMessage, error) {
	r.info.NextAddress = request.GetInfo().NextAddress
	r.info.NextHash = request.GetInfo().NextHash
	r.info.PrevAddress = request.GetInfo().PrevAddress
	r.info.PrevHash = request.GetInfo().PrevHash
	// Let the heartbeat service to deal with this change.
	event.Publish("change", struct{}{})
	return &pb.EmptyMessage{}, nil
}

// MoveGuide helps nodes or files to find their position in the ring. When the
// node and file hash is smaller than the successor node and larger than the
// current node, it can enter the ring.
func (r Router) MoveGuide(ctx context.Context, request *pb.InfoRequest) (*pb.InfoReply, error) {
	if r.info.NextAddress == r.info.CurrAddress {
		return &pb.InfoReply{Info: r.info}, nil
	} else {
		curr := r.info.CurrHash
		next := r.info.NextHash
		// The special case of ring endings is to be avoided here.
		if curr < request.GetInfo().CurrHash && request.GetInfo().CurrHash < next || curr > next {
			return &pb.InfoReply{Info: r.info}, nil
		}
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(r.info.NextAddress, opts...)
		if err != nil {
			return nil, err
		}
		client := pb.NewRouterClient(conn)
		// Recursively call the next node to see if the conditions are met.
		return client.MoveGuide(context.Background(), request)
	}
}
