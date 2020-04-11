package vsphere

import (
	"io"

	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

type vsphereManagerRest struct {
	clusterName string
}

// configFile is used to read and store information from the cloud configuration file
type configFile struct {
	clusterName string `gcfg:"cluster-name"`
}

// createVsphereManagerRest sets up the client and returns
// a vsphereManagerRest
func createVsphereManagerRest(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*vsphereManagerRest, error) {
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

	manager := &vsphereManagerRest{
		clusterName: cfg.clusterName,
	}
	return manager, nil
}

func (mgr *vsphereManagerRest) nodeGroupSize(nodegroup string) (int, error) {
	return 0, cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManagerRest) createNodes(nodegroup string, nodes int) error {
	return cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManagerRest) getNodes(nodegroup string) ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}
	
func (mgr *vsphereManagerRest) getNodeNames(nodegroup string) ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManagerRest) deleteNodes(nodegroup string, nodes []nodeRef, updatedNodeCount int) error {
	return cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManagerRest) templateNodeInfo(nodegroup string) (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
