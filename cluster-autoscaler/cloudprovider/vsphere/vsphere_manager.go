package vsphere

import (
	"io"

	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
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

// configFile is used to read and store information from the cloud configuration file
type configFile struct {
	clusterName string `gcfg:"cluster-name"`
}

type vsphereManager struct {
	clusterName string
}

func newVsphereManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*vsphereManager, error) {
	var cfg configFile
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	if opts.ClusterName == "" && cfg.clusterName == "" {
		klog.Fatalf("The cluster-name parameter must be set")
	} else if opts.ClusterName != "" && cfg.clusterName == "" {
		cfg.clusterName = opts.ClusterName
	}

	manager := &vsphereManager{
		clusterName: cfg.clusterName,
	}
	return manager, nil
}

func (mgr *vsphereManager) nodeGroupSize(nodegroup string) (int, error) {
	return 0, cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManager) createNodes(nodegroup string, nodes int) error {
	return cloudprovider.ErrNotImplemented
}

// getNodes should return VM ID for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *vsphereManager) getNodes(nodegroup string) ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}
	
func (mgr *vsphereManager) getNodeNames(nodegroup string) ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManager) deleteNodes(nodegroup string, nodes []nodeRef, updatedNodeCount int) error {
	return cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManager) templateNodeInfo(nodegroup string) (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
