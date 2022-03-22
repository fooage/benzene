package service

import (
	"log"
	"net"

	"github.com/spf13/viper"
)

type ServiceInfo struct {
	Name string
	Addr net.Addr
	// there's a extra field
	Extra map[string]interface{}
}

func NewServiceInfo(name string, addr string) *ServiceInfo {
	tcp, err := net.ResolveTCPAddr("tcp", viper.GetString("service.address"))
	if err != nil {
		log.Fatalf("Resolve service addr error: %v\n", err)
		return nil
	}
	return &ServiceInfo{
		Name: viper.GetString("service.name"),
		Addr: tcp,
	}
}

var (
	Info   ServiceInfo
	Guider ServiceGuider
)

// ServiceGuider is a interface for extension service registry.
type ServiceGuider interface {
	Register(ServiceInfo) error
	Deregister(ServiceInfo) error
	PickPeer(string, ServiceInfo) (net.Addr, error)
}

// Hash function used for consistent hashing.
type HashFunc func(data []byte) uint32
