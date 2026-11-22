// Package azure manages one private VM and one retained disk per Cueia workspace.
package azure

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v2"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"golang.org/x/crypto/ssh"
)

//go:embed template.json
var template json.RawMessage

type Config struct {
	Subscription, Group, Region, Subnet, Image, Size, Deployment, AdminPublicKey string
	FileClientKey                                                                ssh.Signer
	HostCA                                                                       ssh.PublicKey
}

type Provider struct {
	config      Config
	resources   *armresources.Client
	deployments *armresources.DeploymentsClient
}

var _ provider.Provider = (*Provider)(nil)

func New(c Config, cred azcore.TokenCredential) (*Provider, error) {
	if c.Subscription == "" || c.Group == "" || c.Region == "" || c.Subnet == "" || c.Image == "" || c.Size == "" || c.Deployment == "" || c.AdminPublicKey == "" || c.FileClientKey == nil || c.HostCA == nil {
		return nil, errors.New("incomplete Azure workspace configuration")
	}
	res, err := armresources.NewClient(c.Subscription, cred, nil)
	if err != nil {
		return nil, err
	}
	dep, err := armresources.NewDeploymentsClient(c.Subscription, cred, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{c, res, dep}, nil
}

func (p *Provider) name(o provider.Owner) (string, error) {
	if o.Workspace == "" || o.Deployment != p.config.Deployment {
		return "", fs.ErrPermission
	}
	sum := sha256.Sum256([]byte(o.Deployment + "\x00" + o.Workspace))
	return fmt.Sprintf("cueia-%x", sum[:16]), nil
}
func (p *Provider) id(typ, name string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s", p.config.Subscription, p.config.Group, typ, name)
}

const computeVersion = "2024-11-01"
const networkVersion = "2024-05-01"

type vmView struct {
	VMID              string
	ProvisioningState string
	StorageProfile    struct {
		DataDisks []struct{ ManagedDisk struct{ ID string } }
	}
	NetworkProfile struct{ NetworkInterfaces []struct{ ID string } }
}
type observation struct {
	machine provider.Machine
	nic     string
	etag    string
}

func properties(value any, dest any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
func (p *Provider) observe(ctx context.Context, name string) (observation, error) {
	var result observation
	var raw *http.Response
	resp, err := p.resources.GetByID(policy.WithCaptureResponse(ctx, &raw), p.id("Microsoft.Compute/virtualMachines", name), computeVersion, nil)
	if err != nil {
		return result, err
	}
	var vm vmView
	if err = properties(resp.Properties, &vm); err != nil {
		return result, err
	}
	if toValue(resp.Tags["cueia-owner"]) != name || len(vm.StorageProfile.DataDisks) != 1 || len(vm.NetworkProfile.NetworkInterfaces) != 1 {
		return result, fs.ErrPermission
	}
	disk := p.id("Microsoft.Compute/disks", name+"-data")
	nic := p.id("Microsoft.Network/networkInterfaces", name+"-nic")
	if !strings.EqualFold(vm.StorageProfile.DataDisks[0].ManagedDisk.ID, disk) || !strings.EqualFold(vm.NetworkProfile.NetworkInterfaces[0].ID, nic) {
		return result, fs.ErrPermission
	}
	if vm.VMID == "" || vm.ProvisioningState != "Succeeded" {
		return result, errors.New("Azure VM is not ready")
	}
	result.machine = provider.Machine{ID: vm.VMID, DiskID: disk}
	result.nic = nic
	result.etag = raw.Header.Get("ETag")
	return result, nil
}
func toValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (p *Provider) FindOrAllocate(ctx context.Context, o provider.Owner) (provider.Machine, error) {
	name, err := p.name(o)
	if err != nil {
		return provider.Machine{}, err
	}
	obs, err := p.observe(ctx, name)
	var apiErr *azcore.ResponseError
	if err == nil || !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		return obs.machine, err
	}
	c := p.config
	params := map[string]*armresources.DeploymentParameter{}
	for key, v := range map[string]string{"name": name, "region": c.Region, "size": c.Size, "image": c.Image, "subnet": c.Subnet, "adminKey": c.AdminPublicKey} {
		params[key] = &armresources.DeploymentParameter{Value: v}
	}
	dep := armresources.Deployment{Properties: &armresources.DeploymentProperties{Mode: to.Ptr(armresources.DeploymentModeIncremental), Template: template, Parameters: params}}
	poll, err := p.deployments.BeginCreateOrUpdate(ctx, c.Group, name, dep, nil)
	if err != nil {
		return provider.Machine{}, err
	}
	if _, err = poll.PollUntilDone(ctx, nil); err != nil {
		return provider.Machine{}, err
	}
	obs, err = p.observe(ctx, name)
	return obs.machine, err
}

func (p *Provider) Replace(ctx context.Context, o provider.Owner, old provider.Machine) (provider.Machine, error) {
	name, err := p.name(o)
	if err != nil {
		return provider.Machine{}, err
	}
	obs, err := p.observe(ctx, name)
	if err != nil {
		return provider.Machine{}, err
	}
	if obs.machine != old {
		return provider.Machine{}, fs.ErrPermission
	}
	if obs.etag == "" {
		return provider.Machine{}, errors.New("Azure VM has no replacement precondition")
	}
	conditional := policy.WithHTTPHeader(ctx, http.Header{"If-Match": []string{obs.etag}})
	poll, err := p.resources.BeginDeleteByID(conditional, p.id("Microsoft.Compute/virtualMachines", name), computeVersion, nil)
	if err != nil {
		return provider.Machine{}, err
	}
	if _, err = poll.PollUntilDone(ctx, nil); err != nil {
		return provider.Machine{}, err
	}
	// Require the retained disk to exist before applying a replacement deployment.
	disk, err := p.resources.GetByID(ctx, old.DiskID, "2024-03-02", nil)
	if err != nil {
		return provider.Machine{}, err
	}
	if toValue(disk.Tags["cueia-owner"]) != name {
		return provider.Machine{}, fs.ErrPermission
	}
	return p.FindOrAllocate(ctx, o)
}
