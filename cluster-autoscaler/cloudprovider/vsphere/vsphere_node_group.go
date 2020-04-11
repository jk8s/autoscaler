package vsphere

type vsphereNodeGroup struct{}

func newVsphereNodeGroup() (*vsphereNodeGroup, error) {
	return &vsphereNodeGroup{}, nil
}
