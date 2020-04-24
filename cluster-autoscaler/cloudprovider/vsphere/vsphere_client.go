package vsphere

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/klog"
)

// Contain is helper function to determine of a node is part of nodeGroups
func Contain(node mo.Reference, nodeGroups []mo.Reference) bool {
	for _, ng := range nodeGroups {
		if node == ng {
			return true
		}
	}
	return false
}

// VsphereClient is the client connection manager which
// holds connections to various API endpoints we need to interface
// with, and REST SDK through alternative libraries
type VsphereClient struct {
	// VIM/govmomi client
	vimClient *govmomi.Client

	// REST client used for tags
	restClient *rest.Client
}

func (c *VsphereClient) TagsManager() (*tags.Manager, error) {
	return tags.NewManager(c.restClient), nil
}

func (c *VsphereClient) ContainObjects(parentObjects, childObjects []mo.Reference) ([]mo.Reference) {
	var containedObjects []mo.Reference
	for _, c := range childObjects {
		if Contain(c, parentObjects) {
			containedObjects = append(containedObjects, c)
		}
	}
	return containedObjects	
}

func (c *VsphereClient) GetTagID(tag string) string {
	tm, _ := c.TagsManager()
	tags, _ := tm.GetTags(context.TODO())
	for _, t := range tags {
		if t.Name == tag {
			return t.ID
		}
	}
	return ""
}

func (c *VsphereClient) GetObjectsFromTag(tag string) []mo.Reference {
	tm, _ := c.TagsManager()
	tagID := c.GetTagID(tag)
	objects, _ := tm.ListAttachedObjects(context.TODO(), tagID)
	return objects
}

func (c *VsphereClient) GetVirtualMachineObjectUUID(obj mo.Reference) (string, error) {
	vm := object.NewVirtualMachine(c.vimClient.Client, obj.Reference())

	var o mo.VirtualMachine
	err := vm.Common.Properties(context.TODO(), vm.Reference(), []string{"config.uuid"}, &o)
	if err != nil {
		return "", err
	}
	return o.Config.Uuid, nil
}

func (c *VsphereClient) CreateVirtualMachine(name, datacenter, resourcePool, template string) error {
	finder := find.NewFinder(c.vimClient.Client, false)

	dc, err := finder.DatacenterOrDefault(context.TODO(), datacenter)
	if err != nil {
		return err
	}

	finder.SetDatacenter(dc)

	pool, err := finder.ResourcePoolOrDefault(context.TODO(), resourcePool)
	if err != nil {
		return err
	}

	folders, err := dc.Folders(context.TODO())
	if err != nil {
		return err
	}

	vm, err := finder.VirtualMachine(context.TODO(), template)
	if err != nil {
		return err
	}

	klog.Infof("Cloning %s to %s...", vm.Reference(), name)

	ref := pool.Reference()

	spec := types.VirtualMachineCloneSpec{
		Location: types.VirtualMachineRelocateSpec{
			Pool: &ref,
		},
	}

	task, err := vm.Clone(context.TODO(), folders.VmFolder, name, spec)
	if err != nil {
		return err
	}

	err = task.Wait(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

// Config holds the vsphere client configuration
type Config struct {
	InsecureFlag bool
	User string
	Password string
	VsphereServer string
}

// NewConfig returns a new config
func NewConfig(user, password, vsphereServer string, insecureFlag bool) (*Config, error) {
	if vsphereServer == "" {
		return nil, fmt.Errorf("vsphere-server must be provided")
	}
	if user == "" {
		return nil, fmt.Errorf("vsphere-user must be provided")
	}
	if password == "" {
		return nil, fmt.Errorf("vsphere-password must be provided")
	}
	c := &Config{
		InsecureFlag: insecureFlag,
		User: user,
		Password: password,
		VsphereServer: vsphereServer,
	}
	return c, nil
}

// vimURL return URL to pass to VIM SOAP client
func (c *Config) vimURL() (*url.URL, error) {
	u, err := url.Parse("https://" + c.VsphereServer + "/sdk")
	if err != nil {
		return nil, fmt.Errorf("Error parse url: %s", err)
	}

	u.User = url.UserPassword(c.User, c.Password)

	return u, nil
}

func (c *Config) Client() (*VsphereClient, error) {
	client := &VsphereClient{}

	u, err := c.vimURL()
	if err != nil {
		return nil, fmt.Errorf("Error generating SOAP endpoint url: %s", err)
	}

	client.vimClient, err = newVimClient(context.TODO(), u, c.InsecureFlag)
	if err != nil {
		return nil, err
	}
	klog.Infof("vsphere client configured for URL: %s", c.VsphereServer)

	client.restClient, err = newRestClient(context.TODO(), client.vimClient, c.User, c.Password)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newVimClient(ctx context.Context, u *url.URL, insecure bool) (*govmomi.Client, error) {
	soapClient := soap.NewClient(u, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	// Only login if the URL contains user information
	if u.User != nil {
		err = c.Login(ctx, u.User)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func newRestClient(ctx context.Context, vimClient *govmomi.Client, user, password string) (*rest.Client, error) {
	c := rest.NewClient(vimClient.Client)
	err := c.Login(ctx, url.UserPassword(user, password))
	if err != nil {
		return nil, fmt.Errorf("Error login rest client")
	}
	return c, nil
}
