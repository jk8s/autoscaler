// +build vsphere

package vsphere

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vsphere"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

var AvailableCloudProviders = []string{
	vsphere.ProviderName,
}

const DefaultCloudProvider = vsphere.ProviderName

func buildCloudProvider(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	switch opts.CloudProviderName {
	case vsphere.ProviderName:
		return vsphere.BuildVsphere(opts, do, rl)
	}
	return nil
}
