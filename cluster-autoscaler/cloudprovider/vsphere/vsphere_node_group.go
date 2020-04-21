package vsphere

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	supportScaleToZero = false
)

// vsphereNodeGroup implements NodeGroup interface from cluster-autoscaler/cloudprovider
//
// A NodeGroup represents a homogenous collection of nodes within a cluster,
// which can be dynamically resized between a minimum and maximum
// number of nodes
type vsphereNodeGroup struct{
	vsphereManager vsphereManager
	id string

	minSize int
	maxSize int
	targetSize *int
}

func (ng *vsphereNodeGroup) MaxSize() int {
	return ng.maxSize
}

func (ng *vsphereNodeGroup) MinSize() int {
	return ng.minSize
}

func (ng *vsphereNodeGroup) TargetSize() (int, error) {
	return *ng.targetSize, nil
}

func (ng *vsphereNodeGroup) IncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

func (ng *vsphereNodeGroup) DeleteNodes([]*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

func (ng *vsphereNodeGroup) DecreaseTargetSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

func (ng *vsphereNodeGroup) Id() string {
	return ng.id
}

func (ng *vsphereNodeGroup) Debug() string {
	return ""
}

func (ng *vsphereNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes, err := ng.vsphereManager.getNodes(ng.id)
	if err != nil {
		return nil, fmt.Errorf("could not get nodes: %v", err)
	}
	var instances []cloudprovider.Instance
	for _, node := range nodes {
		instances = append(instances, cloudprovider.Instance{Id: node})
	}
	return instances, nil
}

func (ng *vsphereNodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (ng *vsphereNodeGroup) Exist() bool {
	return true
}

func (ng *vsphereNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

func (ng *vsphereNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (ng *vsphereNodeGroup) Autoprovisioned() bool {
	return false
}
