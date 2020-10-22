package types

import (
	"github.com/Juniper/contrail-go-api"
	//contrailTypes "github.com/Juniper/contrail-go-api/types"
	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
)

// AnalyticsNode struct defines Contrail Analytics node
type AnalyticsNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// Create creates a AnalyticsNode instance
func (c *AnalyticsNode) Create(nodeList []*AnalyticsNode, nodeName string, contrailClient *contrail.Client) error {
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
func (c *AnalyticsNode) Update(nodeList []*AnalyticsNode, nodeName string, contrailClient *contrail.Client) error {
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List("analytics-node")
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult("analytics-node", &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.AnalyticsNode)
				if typedNode.GetName() == nodeName {
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
func (c *AnalyticsNode) Delete(nodeName string, contrailClient *contrail.Client) error {
	vncNodeList, err := contrailClient.List("analytics-node")
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult("analytics-node", &vncNode)
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
