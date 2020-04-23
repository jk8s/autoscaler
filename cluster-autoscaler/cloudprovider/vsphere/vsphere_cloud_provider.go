package vsphere

import (
	"io"
	"os"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for vsphere
	ProviderName = "vsphere"
)

// vsphereCloudProvider implements CloudProvider interface from cluster-autoscaler module
type vsphereCloudProvider struct {
	vsphereManager  *vsphereManager
	resourceLimiter *cloudprovider.ResourceLimiter
	nodeGroups      []vsphereNodeGroup
}

func newVsphereCloudProvider(vsphereManager *vsphereManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	vcp := &vsphereCloudProvider{
		vsphereManager:  vsphereManager,
		resourceLimiter: resourceLimiter,
		nodeGroups:      []vsphereNodeGroup{},
	}
	return vcp, nil
}

// Name returns the name of cloud provider
func (vcp *vsphereCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups managed by the cloud provider
func (vcp *vsphereCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, len(vcp.nodeGroups))
	for i, group := range vcp.nodeGroups {
		groups[i] = &group
	}
	return groups
}

// AddNodeGroup appends a node group to the list of node groups managed by this cluster provider
func (vcp *vsphereCloudProvider) AddNodeGroup(group vsphereNodeGroup) {
	vcp.nodeGroups = append(vcp.nodeGroups, group)
}

// NodeGroupForNode returns the node group that a given node belongs to.
// Only single node group is currently supported, the first node group is always returned.
func (vcp *vsphereCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; found {
		return nil, nil
	}
	return &(vcp.nodeGroups[0]), nil
}

// Pricing is not implemented
func (vcp *vsphereCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes is not implemented
func (vcp *vsphereCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup is not implemented
func (vcp *vsphereCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns resource constraints for the cloud provider
func (vcp *vsphereCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return vcp.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource
func (vcp *vsphereCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes returns all available GPU types the cloud provider support
func (vcp *vsphereCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// Cleanup currently does nothing
func (vcp *vsphereCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every autoscaler main loop,
// currently prints debug info only
func (vcp *vsphereCloudProvider) Refresh() error {
	for _, ng := range vcp.nodeGroups {
		klog.V(3).Info(ng.Debug())
	}
	return nil
}

// BuildVsphere is called by the autoscaler to build a vsphere cloud provider
func BuildVsphere(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser

	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := newVsphereManager(config, do, opts)
	if err != nil {
		klog.Fatalf("Failed to create vsphere manager: %v", err)
	}

	provider, err := newVsphereCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create vsphere cloud provider: %v", err)
	}

	if len(do.NodeGroupSpecs) == 0 {
		klog.Fatalf("Must specify at least one node group with --nodes=<min>:<max>:<name>,...")
	}

	if len(do.NodeGroupSpecs) > 1 {
		klog.Fatalf("Vsphere autoscaler only supports a single nodegroup for now")
	}

	clusterUpdateLock := sync.Mutex{}

	for _, nodeGroupSpec := range do.NodeGroupSpecs {
		spec, err := dynamic.SpecFromString(nodeGroupSpec, supportScaleToZero)
		if err != nil {
			klog.Fatalf("Could not parse node group sepc %s: %v", nodeGroupSpec, err)
		}

		ng := vsphereNodeGroup{
			vsphereManager: *manager,
			id: spec.Name,
			clusterUpdateMutex: &clusterUpdateLock,
			minSize: spec.MinSize,
			maxSize: spec.MaxSize,
			targetSize: new(int),
		}
		*ng.targetSize, err = ng.vsphereManager.nodeGroupSize(ng.id)
		if err != nil {
			klog.Fatalf("Could not set current nodes in node group: %v", err)
		}
		klog.Infof("Found node group %s with size of %d nodes", ng.id, *ng.targetSize)

		provider.(*vsphereCloudProvider).AddNodeGroup(ng)
	}

	return provider
}
