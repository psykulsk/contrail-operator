package databasenodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
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
	databaseInfoLog = log.New(os.Stdout, "databasenodes: ", log.LstdFlags)
}

// Create creates a DatabaseNode instance
func (c *DatabaseNode) Create(nodeList []*DatabaseNode, nodeName string, contrailClient types.ApiClient) error {
	databaseInfoLog.Println("Creating " + c.Hostname + " " + databaseNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.DatabaseNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetDatabaseNodeIpAddress(node.IPAddress)
			annotations := &contrailTypes.KeyValuePairs{
				KeyValuePair: types.ConvertMapToContrailKeyValuePairs(node.Annotations),
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
func (c *DatabaseNode) Update(nodeList []*DatabaseNode, nodeName string, contrailClient types.ApiClient) error {
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
					if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
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
func (c *DatabaseNode) Delete(nodeName string, contrailClient types.ApiClient) error {
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
			if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
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

func ReconcileDatabaseNodes(contrailClient types.ApiClient, nodeList []*DatabaseNode) error {
	var actionMap = make(map[string]string)
	nodeType := "database-node"
	vncNodes := []*DatabaseNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	databaseInfoLog.Printf("VncNodeList: %v\n", vncNodeList)
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.DatabaseNode)

		node := &DatabaseNode{
			IPAddress: typedNode.GetDatabaseNodeIpAddress(),
			Hostname:  typedNode.GetName(),
		}
		vncNodes = append(vncNodes, node)
	}
	for _, node := range nodeList {
		actionMap[node.Hostname] = "create"
	}
	databaseInfoLog.Printf("VncNodes: %v\n", vncNodes)

	for _, vncNode := range vncNodes {
		if _, ok := actionMap[vncNode.Hostname]; ok {
			for _, node := range nodeList {
				if node.Hostname == vncNode.Hostname {
					actionMap[node.Hostname] = "noop"
					if node.IPAddress != vncNode.IPAddress {
						actionMap[node.Hostname] = "update"
					}
				}
			}
		} else {
			actionMap[vncNode.Hostname] = "delete"
		}
	}
	for k, v := range actionMap {
		databaseInfoLog.Printf("actionMapValue: %v\n", v)
		switch v {
		case "update":
			for _, node := range nodeList {
				if node.Hostname == k {
					databaseInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					databaseInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			node := &DatabaseNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
			databaseInfoLog.Println("deleting node ", k)
		}
	}
	return nil
}
