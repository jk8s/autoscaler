package vsphere

import (
	"fmt"
	"io"
	"os"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	defaultManager = "rest"
)

// NodeRef stores name, machineID and providerID of a node
type nodeRef struct {
	name string
	machineID string
	providerID string
	ips []string
}

// vsphereManager is an interface for basic interaction with the cluster
type vsphereManager interface{
	nodeGroupSize(nodegroup string) (int, error)
	createNodes(nodegroup string, nodes int) error
	getNodes(nodegroup string) ([]string, error)
	getNodeNames(nodegroup string) ([]string, error)
	deleteNodes(nodegroup string, nodes []nodeRef, updatedNodeCount int) error
	templateNodeInfo(nodegroup string) (*schedulernodeinfo.NodeInfo, error)
}

func createVsphereManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (vsphereManager, error) {
	// Get manager from env var, options: rest, vim, pbm
	// Ref: https://github.com/terraform-providers/terraform-provider-vsphere/blob/master/vsphere/config.go#L33
	manager, ok := os.LookupEnv("VSPHERE_MANAGER")
	if !ok {
		manager = defaultManager
	}

	switch manager {
	case "rest":
		return createVsphereManagerRest(configReader, discoverOpts, opts)
	}
	return nil, fmt.Errorf("vsphere manager does not exist: %s", manager)
}
