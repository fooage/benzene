package main

import (
	"log"
	"net"

	"github.com/fooage/messier/proto/pb"
	"github.com/fooage/messier/service/connect"
	"github.com/fooage/messier/service/keepalive"
	"github.com/fooage/messier/service/transport"
	"github.com/fooage/messier/utils"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	opts        []grpc.ServerOption
	local       string // local address for this node
	access      string // address which is accessed
	information pb.Information
)

// Initialize node information from the configuration file.
func init() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("read config occured an error: %v\n", err)
	}
	// TODO: Set some options of server.
	local = viper.GetString("server.connect.local_address")
	access = viper.GetString("server.connect.access_address")
	information = pb.Information{
		CurrAddress: local,
		CurrHash:    utils.EncodeStringHash(local),
		NextAddress: local,
		NextHash:    utils.EncodeStringHash(local),
		PrevAddress: local,
		PrevHash:    utils.EncodeStringHash(local),
	}
}

func main() {
	lis, err := net.Listen("tcp", local)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	server := grpc.NewServer(opts...)
	router := connect.NewRouter(&information)
	heartbeat := keepalive.NewHeartbeat(&information)
	porter := transport.NewPorter(&information)
	// TODO: Add the connection logic code for router and heartbeat service.
	// If this node is original node in the cluster, it will not access other
	// node. Key goroutines for connection and disconnection and reconnection
	// in the cluster's ring.
	go router.ConnectCluster(access)
	go heartbeat.HeartKeeping()
	go porter.AdjustStorage()
	pb.RegisterRouterServer(server, router)
	pb.RegisterHeartbeatServer(server, heartbeat)
	pb.RegisterPorterServer(server, porter)
	if err = server.Serve(lis); err != nil {
		log.Fatalf("server happend an error: %v\n", err)
	}
}
