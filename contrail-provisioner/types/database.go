package types

import (
	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
)

// DatabaseNode struct defines Contrail database node
type DatabaseNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// Create creates a DatabaseNode instance
func (c *DatabaseNode) Create(nodeList []*DatabaseNode, nodeName string, contrailClient ApiClient) error {
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.DatabaseNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetDatabaseNodeIpAddress(node.IPAddress)
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

// Update updates a DatabaseNode instance
func (c *DatabaseNode) Update(nodeList []*DatabaseNode, nodeName string, contrailClient ApiClient) error {
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List("database-node")
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult("database-node", &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.DatabaseNode)
				if typedNode.GetName() == nodeName {
					typedNode.SetFQName("", []string{"default-global-system-config", nodeName})
					typedNode.SetDatabaseNodeIpAddress(node.IPAddress)
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

// Delete deletes a DatabaseNode instance
func (c *DatabaseNode) Delete(nodeName string, contrailClient ApiClient) error {
	vncNodeList, err := contrailClient.List("database-node")
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult("database-node", &vncNode)
		if err != nil {
			return err
		}
		if obj.GetName() == nodeName {
			err = contrailClient.Delete(obj)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
