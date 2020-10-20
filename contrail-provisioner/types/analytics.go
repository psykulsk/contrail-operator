package types

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
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
	analyticsInfoLog = log.New(os.Stdout, "analytics: ", log.LstdFlags)
}

// Create creates a AnalyticsNode instance
func (c *AnalyticsNode) Create(nodeList []*AnalyticsNode, nodeName string, contrailClient ApiClient) error {
	analyticsInfoLog.Println("Creating " + c.Hostname + " " + analyticsNodeType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.AnalyticsNode{}
			vncNode.SetFQName("", []string{"default-global-system-config", nodeName})
			vncNode.SetAnalyticsNodeIpAddress(node.IPAddress)
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

// Update updates a AnalyticsNode instance
func (c *AnalyticsNode) Update(nodeList []*AnalyticsNode, nodeName string, contrailClient ApiClient) error {
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
					if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
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
func (c *AnalyticsNode) Delete(nodeName string, contrailClient ApiClient) error {
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
			if !HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
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
