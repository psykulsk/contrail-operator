package types

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
)

// ConfigNode struct defines Contrail config node
type ConfigNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const configNodeType string = "config-node"

var configInfoLog *log.Logger

func init() {
	configInfoLog = log.New(os.Stdout, "config: ", log.LstdFlags)
}

// Create creates a ConfigNode instance
func (c *ConfigNode) Create(nodeList []*ConfigNode, nodeName string, contrailClient ApiClient) error {
	configInfoLog.Println("Creating " + c.Hostname + " " + configNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.ConfigNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetConfigNodeIpAddress(node.IPAddress)
			annotations := &contrailTypes.KeyValuePairs{
				KeyValuePair: ConvertMapToContrailKeyValuePairs(node.Annotations),
			}
			vncNode.SetAnnotations(annotations)
			err := contrailClient.Create(vncNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Update updates a ConfigNode instance
func (c *ConfigNode) Update(nodeList []*ConfigNode, nodeName string, contrailClient ApiClient) error {
	configInfoLog.Println("Updating " + c.Hostname + " " + configNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List(configNodeType)
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult(configNodeType, &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.ConfigNode)
				if typedNode.GetName() == nodeName {
					if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
						configInfoLog.Println(c.Hostname + " " + configNodeType + " does not have the required annotations.")
						configInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + configNodeType)
						return nil
					}
					typedNode.SetFQName("", []string{"default-global-system-config", nodeName})
					typedNode.SetConfigNodeIpAddress(node.IPAddress)
					err := contrailClient.Update(typedNode)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Delete deletes a ConfigNode instance
func (c *ConfigNode) Delete(nodeName string, contrailClient ApiClient) error {
	configInfoLog.Println("Deleting " + c.Hostname + " " + configNodeType)
	vncNodeList, err := contrailClient.List(configNodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(configNodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.ConfigNode)
		if typedNode.GetName() == nodeName {
			if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
				configInfoLog.Println(c.Hostname + " " + configNodeType + " does not have the required annotations.")
				configInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + configNodeType)
				return nil
			}
			err = contrailClient.Delete(obj)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
