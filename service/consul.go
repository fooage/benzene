package service

import (
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
)

// ConsulGuider is responsible for completing the interaction with consul.
type ConsulGuider struct {
	addr net.Addr
	// keep a long connection with consul
	client *api.Client
	hash   HashFunc
}

func NewConsulGuider(addr string, consistent string) *ConsulGuider {
	target, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("Trying to resolve consul addr error: %v\n", err)
		return nil
	}
	config := api.DefaultConfig()
	config.Address = target.String()
	// set the default hash func
	var hash HashFunc
	switch consistent {
	case "crc32":
		hash = crc32.ChecksumIEEE
	case "adler32":
		hash = adler32.Checksum
	default:
		hash = crc32.ChecksumIEEE
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Can not connect with consul: %v\n", err)
		return nil
	}
	return &ConsulGuider{
		addr:   target,
		client: client,
		hash:   hash,
	}
}

func (c *ConsulGuider) Register(info ServiceInfo) error {
	// register this service in consul
	addr := strings.Split(info.Addr.String(), ":")[0]
	port, _ := strconv.Atoi(strings.Split(info.Addr.String(), ":")[1])
	if err := c.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    info.Name,
		Address: addr,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s", c.addr.String()),
			Timeout:                        "2s",
			Interval:                       "2s",
			DeregisterCriticalServiceAfter: "5s",
		},
	}); err != nil {
		log.Fatalf("Consul register service error: %v\n", err)
		return err
	}
	return nil
}

func (c *ConsulGuider) Deregister(info ServiceInfo) error {
	// deregister this service in consul
	err := c.client.Agent().ServiceDeregister(info.Name)
	if err != nil {
		log.Fatalf("Consul deregister service error: %v\n", err)
		return err
	}
	return nil
}

// SortInstance will use the hash of instance addr to sort, and this is the
// most important part of consistent hash.
type SortInstance struct {
	Addr net.Addr
	Hash uint32
}

func (c *ConsulGuider) PickPeer(item string, info ServiceInfo) (net.Addr, error) {
	catalog, _, err := c.client.Catalog().Service(info.Name, "", nil)
	if err != nil {
		log.Printf("Find service instance error: %v\n", err)
	}
	// TODO: Storage instance info for faster pick.
	ins := make([]SortInstance, len(catalog))
	for idx, val := range catalog {
		addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", val.ServiceAddress, val.ServicePort))
		ins[idx].Addr = addr
		ins[idx].Hash = c.hash([]byte(addr.String()))
	}
	sort.Slice(ins, func(i, j int) bool {
		return ins[i].Hash < ins[j].Hash
	})
	// search the node should be access
	hash := c.hash([]byte(item))
	idx := sort.Search(len(ins), func(i int) bool {
		return ins[i].Hash >= hash
	})
	return ins[idx%len(ins)].Addr, nil
}
