package analyticsnodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/contrailclient"
)

// AnalyticsNode struct defines Contrail Analytics node
type AnalyticsNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const analyticsNodeType string = "analytics-node"

var analyticsInfoLog *log.Logger

func init() {
	analyticsInfoLog = log.New(os.Stdout, "analyticsnodes: ", log.LstdFlags)
}

// Create creates a AnalyticsNode instance
func (c *AnalyticsNode) Create(nodeList []*AnalyticsNode, nodeName string, contrailClient contrailclient.ApiClient) error {
	analyticsInfoLog.Println("Creating " + c.Hostname + " " + analyticsNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.AnalyticsNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetAnalyticsNodeIpAddress(node.IPAddress)
			annotations := contrailclient.ConvertMapToContrailKeyValuePairs(node.Annotations)
			vncNode.SetAnnotations(&annotations)
			err := contrailClient.Create(vncNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Update updates a AnalyticsNode instance
func (c *AnalyticsNode) Update(nodeList []*AnalyticsNode, nodeName string, contrailClient contrailclient.ApiClient) error {
	analyticsInfoLog.Println("Updating " + c.Hostname + " " + analyticsNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List(analyticsNodeType)
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult(analyticsNodeType, &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.AnalyticsNode)
				if typedNode.GetName() == nodeName {
					storedAnnotations := contrailclient.ConvertContrailKeyValuePairsToMap(typedNode.GetAnnotations())
					if !contrailclient.HasRequiredAnnotations(storedAnnotations, c.Annotations) {
						analyticsInfoLog.Println(c.Hostname + " " + analyticsNodeType + " does not have the required annotations.")
						analyticsInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + analyticsNodeType)
						return nil
					}
					typedNode.SetFQName("", []string{"default-global-system-config", nodeName})
					typedNode.SetAnalyticsNodeIpAddress(node.IPAddress)
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

// Delete deletes a AnalyticsNode instance
func (c *AnalyticsNode) Delete(nodeName string, contrailClient contrailclient.ApiClient) error {
	analyticsInfoLog.Println("Deleting " + c.Hostname + " " + analyticsNodeType)
	vncNodeList, err := contrailClient.List(analyticsNodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(analyticsNodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.AnalyticsNode)
		if typedNode.GetName() == nodeName {
			storedAnnotations := contrailclient.ConvertContrailKeyValuePairsToMap(typedNode.GetAnnotations())
			if !contrailclient.HasRequiredAnnotations(storedAnnotations, c.Annotations) {
				analyticsInfoLog.Println(c.Hostname + " " + analyticsNodeType + " does not have the required annotations.")
				analyticsInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + analyticsNodeType)
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

func ReconcileAnalyticsNodes(contrailClient contrailclient.ApiClient, nodeList []*AnalyticsNode) error {
	var actionMap = make(map[string]string)
	nodeType := "analytics-node"
	vncNodes := []*AnalyticsNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.AnalyticsNode)

		node := &AnalyticsNode{
			IPAddress: typedNode.GetAnalyticsNodeIpAddress(),
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
					analyticsInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					analyticsInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			analyticsInfoLog.Println("deleting node ", k)
			node := &AnalyticsNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
