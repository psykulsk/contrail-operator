package confignodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
)

// ConfigNode struct defines Contrail config node
type ConfigNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const nodeType string = "config-node"

var configInfoLog *log.Logger

func init() {
	configInfoLog = log.New(os.Stdout, "confignodes: ", log.LstdFlags)
}

// Create creates a ConfigNode instance
func (c *ConfigNode) Create(nodeList []*ConfigNode, nodeName string, contrailClient types.ApiClient) error {
	configInfoLog.Println("Creating " + c.Hostname + " " + nodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.ConfigNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetConfigNodeIpAddress(node.IPAddress)
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

// Update updates a ConfigNode instance
func (c *ConfigNode) Update(nodeList []*ConfigNode, nodeName string, contrailClient types.ApiClient) error {
	configInfoLog.Println("Updating " + c.Hostname + " " + nodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List(nodeType)
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.ConfigNode)
				if typedNode.GetName() == nodeName {
					if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
						configInfoLog.Println(c.Hostname + " " + nodeType + " does not have the required annotations.")
						configInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + nodeType)
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
func (c *ConfigNode) Delete(nodeName string, contrailClient types.ApiClient) error {
	configInfoLog.Println("Deleting " + c.Hostname + " " + nodeType)
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.ConfigNode)
		if typedNode.GetName() == nodeName {
			if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
				configInfoLog.Println(c.Hostname + " " + nodeType + " does not have the required annotations.")
				configInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + nodeType)
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

func ReconcileConfigNodes(contrailClient types.ApiClient, nodeList []*ConfigNode) error {
	var actionMap = make(map[string]string)
	vncNodes := []*ConfigNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.ConfigNode)

		node := &ConfigNode{
			IPAddress: typedNode.GetConfigNodeIpAddress(),
			Hostname:  typedNode.GetName(),
		}
		vncNodes = append(vncNodes, node)
	}
	for _, node := range nodeList {
		actionMap[node.Hostname] = "create"
	}
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
		switch v {
		case "update":
			for _, node := range nodeList {
				if node.Hostname == k {
					configInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					configInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			node := &ConfigNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
			configInfoLog.Println("deleting node ", k)
		}
	}
	return nil
}
