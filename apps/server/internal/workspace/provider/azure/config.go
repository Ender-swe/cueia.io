package azure

import (
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"golang.org/x/crypto/ssh"
	"os"
)

// FromEnvironment loads the deployment's server identity and private file-service credentials.
func FromEnvironment() (*Provider, error) {
	c := Config{Subscription: os.Getenv("AZURE_SUBSCRIPTION_ID"), Group: os.Getenv("AZURE_RESOURCE_GROUP"), Region: config.AzureLocation(), Subnet: os.Getenv("AZURE_SUBNET_ID"), Image: os.Getenv("AZURE_IMAGE_ID"), Size: os.Getenv("AZURE_VM_SIZE"), Deployment: config.DeploymentID(), AdminPublicKey: os.Getenv("AZURE_VM_ADMIN_PUBLIC_KEY")}
	data, err := os.ReadFile(os.Getenv("CUEIA_FILE_CLIENT_KEY_PATH"))
	if err != nil {
		return nil, errors.New("cannot read CUEIA_FILE_CLIENT_KEY_PATH")
	}
	c.FileClientKey, err = ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, errors.New("invalid file client key")
	}
	ca, err := os.ReadFile(os.Getenv("CUEIA_HOST_CA_PATH"))
	if err != nil {
		return nil, errors.New("cannot read CUEIA_HOST_CA_PATH")
	}
	c.HostCA, _, _, _, err = ssh.ParseAuthorizedKey(ca)
	if err != nil {
		return nil, errors.New("invalid workspace host CA")
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return New(c, cred)
}
