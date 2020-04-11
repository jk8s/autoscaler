package vsphere

type vsphereManager struct{}

func newVsphereManager() (*vsphereManager, error) {
	return &vsphereManager{}, nil
}
