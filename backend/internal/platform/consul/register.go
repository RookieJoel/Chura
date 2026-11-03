package consul

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/consul/api"
)

// Register registers the service with Consul when CONSUL_HTTP_ADDR is set.
// The catalog address is always an IP so Consul DNS returns A records Kong can use.
// Returns a cleanup function to deregister on shutdown, or nil when skipped.
func Register(serviceName string, serviceAddress string, port int, healthPath string) (func(), error) {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		return nil, nil
	}

	ip, err := registerAddress(serviceAddress, "consul:8500")
	if err != nil {
		return nil, fmt.Errorf("consul service address: %w", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = ip
	}
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, hostname, port)

	client, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: ip,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", ip, port, healthPath),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		return nil, fmt.Errorf("consul register: %w", err)
	}

	log.Printf("consul: registered %s as %s (%s:%d)", serviceName, serviceID, ip, port)

	cleanup := func() {
		if err := client.Agent().ServiceDeregister(serviceID); err != nil {
			log.Printf("consul: deregister %s: %v", serviceID, err)
		}
	}
	return cleanup, nil
}
