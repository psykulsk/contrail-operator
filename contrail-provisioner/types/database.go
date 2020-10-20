package types

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
)

// DatabaseNode struct defines Contrail database node
type DatabaseNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const databaseNodeType string = "database-node"

var databaseInfoLog *log.Logger

func init() {
	databaseInfoLog = log.New(os.Stdout, "database: ", log.LstdFlags)
}

// Create creates a DatabaseNode instance
func (c *DatabaseNode) Create(nodeList []*DatabaseNode, nodeName string, contrailClient ApiClient) error {
	databaseInfoLog.Println("Creating " + c.Hostname + " " + databaseNodeType)
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
	databaseInfoLog.Println("Updating " + c.Hostname + " " + databaseNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List(databaseNodeType)
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult(databaseNodeType, &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.DatabaseNode)
				if typedNode.GetName() == nodeName {
					if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
						databaseInfoLog.Println(c.Hostname + " " + databaseNodeType + " does not have the required annotations.")
						databaseInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + databaseNodeType)
						return nil
					}
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
	databaseInfoLog.Println("Deleting " + c.Hostname + " " + databaseNodeType)
	vncNodeList, err := contrailClient.List(databaseNodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(databaseNodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.DatabaseNode)
		if obj.GetName() == nodeName {
			if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
				databaseInfoLog.Println(c.Hostname + " " + databaseNodeType + " does not have the required annotations.")
				databaseInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + databaseNodeType)
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
