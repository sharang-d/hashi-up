package config

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type BoundaryConfig struct {
	WorkerName           string
	ControllerName       string
	DatabaseURL          string
	RootKey              string
	WorkerAuthKey        string
	RecoveryKey          string
	ApiAddress           string
	ClusterAddress       string
	ProxyAddress         string
	PublicAddress        string
	PublicClusterAddress string
	Controllers          []string
}

func (c *BoundaryConfig) IsWorkerEnabled() bool {
	return len(c.WorkerName) != 0
}

func (c *BoundaryConfig) IsControllerEnabled() bool {
	return len(c.ControllerName) != 0
}

func (c *BoundaryConfig) GenerateConfigFile() string {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	if len(c.ControllerName) != 0 {
		controllerBlock := rootBody.AppendNewBlock("controller", []string{})
		controllerBlock.Body().SetAttributeValue("name", cty.StringVal(c.ControllerName))

		databaseBlock := controllerBlock.Body().AppendNewBlock("database", []string{})
		databaseBlock.Body().SetAttributeValue("url", cty.StringVal(c.DatabaseURL))

		if len(c.PublicClusterAddress) != 0 {
			controllerBlock.Body().SetAttributeValue("public_cluster_addr", cty.StringVal(c.PublicClusterAddress))
		}
	}

	if len(c.WorkerName) != 0 {
		workerBlock := rootBody.AppendNewBlock("worker", []string{})
		workerBlock.Body().SetAttributeValue("name", cty.StringVal(c.WorkerName))
		workerBlock.Body().SetAttributeValue("controllers", cty.ListVal(transform(c.Controllers)))

		if len(c.PublicAddress) != 0 {
			workerBlock.Body().SetAttributeValue("public_addr", cty.StringVal(c.PublicAddress))
		}
	}

	if len(c.ControllerName) != 0 && len(c.ApiAddress) != 0 {
		apiAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		apiAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("api"))
		apiAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ApiAddress))
		apiAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
	}

	if len(c.ControllerName) != 0 && len(c.ClusterAddress) != 0 {
		apiAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		apiAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("cluster"))
		apiAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ClusterAddress))
		apiAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
	}

	if len(c.WorkerName) != 0 && len(c.ProxyAddress) != 0 {
		apiAddressBlock := rootBody.AppendNewBlock("listener", []string{"tcp"})
		apiAddressBlock.Body().SetAttributeValue("purpose", cty.StringVal("proxy"))
		apiAddressBlock.Body().SetAttributeValue("address", cty.StringVal(c.ProxyAddress))
		apiAddressBlock.Body().SetAttributeValue("tls_disable", cty.BoolVal(true))
	}

	if len(c.RootKey) != 0 {
		rootKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		rootKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("root"))
		rootKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		rootKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.RootKey))
		rootKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_root"))
	}

	if len(c.WorkerAuthKey) != 0 {
		workerAuthKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		workerAuthKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("worker-auth"))
		workerAuthKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		workerAuthKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.WorkerAuthKey))
		workerAuthKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_worker-auth"))
	}

	if len(c.RecoveryKey) != 0 {
		recoveryKeyBlock := rootBody.AppendNewBlock("kms", []string{"aead"})
		recoveryKeyBlock.Body().SetAttributeValue("purpose", cty.StringVal("recovery"))
		recoveryKeyBlock.Body().SetAttributeValue("aead_type", cty.StringVal("aes-gcm"))
		recoveryKeyBlock.Body().SetAttributeValue("key", cty.StringVal(c.RecoveryKey))
		recoveryKeyBlock.Body().SetAttributeValue("key_id", cty.StringVal("global_recovery"))
	}

	return string(f.Bytes())
}

func generateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
