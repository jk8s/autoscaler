package vsphere

import (
	"fmt"
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

// ConfigVsphere is used to read and store information from the cloud configuration file
type ConfigVsphere struct {
	ClusterName string `gcfg:"cluster-name"`
	VsphereServer string `gcfg:"vsphere-server"`
	VsphereUsername string `gcfg:"vsphere-username"`
	VspherePassword string `gcfg:"vsphere-password"`
	VsphereInsecureFlag bool `gcfg:"vsphere-insecure"`
	VsphereDatacenter string `gcfg:"vsphere-datacenter"`
	VsphereResourcePool string `gcfg:"vsphere-resource-pool"`
	VsphereTemplate string `gcfg:"vsphere-template"`
}

// ConfigFile is used to read and store information from the cloud configuration file
type ConfigFile struct {
	Vsphere ConfigVsphere `gcfg:"vsphere"`
}

// VirtualMachineSpec represents a Vsphere virtual machine
type VirtualMachineSpec struct {
	Tags []string
}

type vsphereManager struct {
	clusterName string
	datacenter string
	resourcePool string
	template string
	vsphereClient *VsphereClient
}

func newVsphereManager(configReader io.Reader, discoverOpts cloudprovider.NodeGroupDiscoveryOptions, opts config.AutoscalingOptions) (*vsphereManager, error) {
	var cfg ConfigFile
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	if opts.ClusterName == "" && cfg.Vsphere.ClusterName == "" {
		klog.Fatalf("The cluster-name parameter must be set")
	} else if opts.ClusterName != "" && cfg.Vsphere.ClusterName == "" {
		cfg.Vsphere.ClusterName = opts.ClusterName
	}

	config, err := NewConfig(cfg.Vsphere.VsphereUsername, cfg.Vsphere.VspherePassword, cfg.Vsphere.VsphereServer, cfg.Vsphere.VsphereInsecureFlag)
	if err != nil {
		klog.Fatalf("Vsphere config is invalid")
		return nil, err
	}

	vsphereClient, err := config.Client()
	if err != nil {
		klog.Fatalf("Failed initializing vsphere client")
		return nil, err
	}

	klog.Infof("Starting vsphere manager with config: %v", cfg)
	manager := &vsphereManager{
		clusterName: cfg.Vsphere.ClusterName,
		datacenter: cfg.Vsphere.VsphereDatacenter,
		resourcePool: cfg.Vsphere.VsphereResourcePool,
		template: cfg.Vsphere.VsphereTemplate,
		vsphereClient: vsphereClient,
	}
	return manager, nil
}

// nodeGroupSize gets the current size of the nodegroup as reported by vsphere tags
func (mgr *vsphereManager) nodeGroupSize(nodeGroup string) (int, error) {
	clusterMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-cluster-"+mgr.clusterName)
	nodeGroupMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-nodegroup-"+nodeGroup)
	nodes := mgr.vsphereClient.ContainObjects(clusterMachines, nodeGroupMachines)
	klog.V(3).Infof("Nodegroup %s: %d/%d", nodeGroup, len(nodes), len(clusterMachines))
	return len(nodes), nil
}

func (mgr *vsphereManager) createNode(name string) error {
	// TODO(giri): Pass cloud-init

	err := mgr.vsphereClient.CreateVirtualMachine(name, mgr.datacenter, mgr.resourcePool, mgr.template)
	if err != nil {
		return err
	}
	klog.Infof("Virtual machine %s has been created", name)
	return nil
}

func (mgr *vsphereManager) createNodes(nodeGroup string, nodes int) error {
	klog.Infof("Updating node count to %d for nodegroup %s", nodes, nodeGroup)
	
	// TODO(giri): Add cloud-init script

	for i := 0; i < nodes; i++ {
		nodeName := fmt.Sprintf("%s-%s-%02d", mgr.clusterName, nodeGroup, i+nodes+1)
		mgr.createNode(nodeName)
	}
	return nil
}

// getNodes should return ProviderIDs (use VM ID as Provider ID) for all nodes in the node group,
// used to find any nodes which are unregistered in kubernetes.
func (mgr *vsphereManager) getNodes(nodeGroup string) ([]string, error) {
	clusterMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-cluster-"+mgr.clusterName)
	nodeGroupMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-nodegroup-"+nodeGroup)
	nodes := mgr.vsphereClient.ContainObjects(clusterMachines, nodeGroupMachines)
	nodeIDs := []string{}
	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.Reference().Value)
	}
	return nodeIDs, nil
}
	
func (mgr *vsphereManager) getNodeNames(nodeGroup string) ([]string, error) {
	clusterMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-cluster-"+mgr.clusterName)
	nodeGroupMachines := mgr.vsphereClient.GetObjectsFromTag("k8s-nodegroup-"+nodeGroup)
	nodes := mgr.vsphereClient.ContainObjects(clusterMachines, nodeGroupMachines)
	nodeIDs := []string{}
	for _, n := range nodes {
		// TODO(giri): Add additional call to get VM hostname
		nodeIDs = append(nodeIDs, n.Reference().Value)
	}
	return nodeIDs, nil
}

func (mgr *vsphereManager) deleteNodes(nodegroup string, nodes []nodeRef, updatedNodeCount int) error {
	return cloudprovider.ErrNotImplemented
}

func (mgr *vsphereManager) templateNodeInfo(nodegroup string) (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
