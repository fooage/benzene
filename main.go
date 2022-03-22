package main

import (
	"log"

	"github.com/fooage/benzene/service"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Read config occured an error: %v\n", err)
	}
	// init this node's service info
	name := viper.GetString("service.name")
	addr := viper.GetString("service.address")
	service.Info = *service.NewServiceInfo(name, addr)
	// init service discovery
	addr = viper.GetString("discovery.address")
	consistent := viper.GetString("discovery.consistent_hash")
	service.Guider = service.NewConsulGuider(addr, consistent)
}

func main() {
	// Register this service with service governance. The registrar needs to
	// satisfy the defined interface include register and deregister.
	registerService(service.Guider)
	defer deregisterService(service.Guider)
	// TODO: To add the file or cache service.

}

func registerService(r service.ServiceGuider) {
	err := r.Register(service.Info)
	if err != nil {
		panic(err)
	}
}

func deregisterService(r service.ServiceGuider) {
	err := r.Deregister(service.Info)
	if err != nil {
		panic(err)
	}
}
