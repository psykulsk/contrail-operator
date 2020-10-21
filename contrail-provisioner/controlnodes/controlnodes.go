package controlnodes

import (
	"log"
	"os"
	"reflect"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
)

// ControlNode struct defines Contrail control node
type ControlNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	ASN         int               `yaml:"asn,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const bgpRouterType string = "bgp-router"

var controlInfoLog *log.Logger

func init() {
	controlInfoLog = log.New(os.Stdout, "controlnodes: ", log.LstdFlags)
}

// Create creates a ControlNode instance
func (c *ControlNode) Create(nodeList []*ControlNode, nodeName string, contrailClient types.ApiClient) error {
	controlInfoLog.Println("Creating " + c.Hostname + " " + bgpRouterType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNode := &contrailTypes.BgpRouter{}
			vncNode.SetFQName("", []string{"default-domain", "default-project", "ip-fabric", "__default__", nodeName})
			vncNode.SetName(nodeName)
			annotations := &contrailTypes.KeyValuePairs{
				KeyValuePair: types.ConvertMapToContrailKeyValuePairs(node.Annotations),
			}
			vncNode.SetAnnotations(annotations)
			bgpParameters := &contrailTypes.BgpRouterParams{
				Address:          node.IPAddress,
				AutonomousSystem: node.ASN,
				Vendor:           "contrail",
				RouterType:       "control-node",
				AdminDown:        false,
				Identifier:       node.IPAddress,
				HoldTime:         90,
				Port:             179,
				AddressFamilies: &contrailTypes.AddressFamilies{
					Family: []string{"route-target", "inet-vpn", "inet6-vpn", "e-vpn", "erm-vpn"},
				},
			}
			vncNode.SetBgpRouterParameters(bgpParameters)

			routingInstance := &contrailTypes.RoutingInstance{}
			routingInstanceObjectsList, err := contrailClient.List("routing-instance")
			if err != nil {
				return err
			}

			if len(routingInstanceObjectsList) == 0 {
				controlInfoLog.Println("no routingInstance objects")
			}

			for _, routingInstanceObject := range routingInstanceObjectsList {
				obj, err := contrailClient.ReadListResult("routing-instance", &routingInstanceObject)
				if err != nil {
					return err
				}
				if reflect.DeepEqual(obj.GetFQName(), []string{"default-domain", "default-project", "ip-fabric", "__default__"}) {
					routingInstance = obj.(*contrailTypes.RoutingInstance)
				}
			}

			if routingInstance != nil {
				vncNode.SetParent(routingInstance)
			}

			err = contrailClient.Create(vncNode)
			if err != nil {
				return err
			}

			gscObjects := []*contrailTypes.GlobalSystemConfig{}
			gscObjectsList, err := contrailClient.List("global-system-config")
			if err != nil {
				return err
			}

			if len(gscObjectsList) == 0 {
				controlInfoLog.Println("no gscObject")
			}

			for _, gscObject := range gscObjectsList {
				obj, err := contrailClient.ReadListResult("global-system-config", &gscObject)
				if err != nil {
					return err
				}
				gscObjects = append(gscObjects, obj.(*contrailTypes.GlobalSystemConfig))
			}

			if len(gscObjects) > 0 {
				for _, gsc := range gscObjects {
					if err := gsc.AddBgpRouter(vncNode); err != nil {
						return err
					}
					if err := contrailClient.Update(gsc); err != nil {
						return err
					}
				}
			}

		}
	}

	gscObjects := []*contrailTypes.GlobalSystemConfig{}
	gscObjectsList, err := contrailClient.List("global-system-config")
	if err != nil {
		return err
	}

	if len(gscObjectsList) == 0 {
		controlInfoLog.Println("no gscObject")
	}

	for _, gscObject := range gscObjectsList {
		obj, err := contrailClient.ReadListResult("global-system-config", &gscObject)
		if err != nil {
			return err
		}
		gscObjects = append(gscObjects, obj.(*contrailTypes.GlobalSystemConfig))
	}

	if len(gscObjects) > 0 {
		for _, gsc := range gscObjects {
			bgpRefs, err := gsc.GetBgpRouterRefs()
			if err != nil {
				return err
			}
			for _, bgpRef := range bgpRefs {
				controlInfoLog.Println(bgpRef)
			}

		}
	}

	return nil
}

// Update updates a ControlNode instance
func (c *ControlNode) Update(nodeList []*ControlNode, nodeName string, contrailClient types.ApiClient) error {
	controlInfoLog.Println("Updating " + c.Hostname + " " + bgpRouterType)
	for _, node := range nodeList {
		if node.Hostname == nodeName {
			vncNodeList, err := contrailClient.List(bgpRouterType)
			if err != nil {
				return err
			}
			for _, vncNode := range vncNodeList {
				obj, err := contrailClient.ReadListResult(bgpRouterType, &vncNode)
				if err != nil {
					return err
				}
				typedNode := obj.(*contrailTypes.BgpRouter)
				if typedNode.GetName() == nodeName {
					if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
						controlInfoLog.Println(c.Hostname + " " + bgpRouterType + " does not have the required annotations.")
						controlInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + bgpRouterType)
						return nil
					}
					typedNode.SetFQName("", []string{"default-domain", "default-project", "ip-fabric", "__default__", nodeName})
					bgpParameters := &contrailTypes.BgpRouterParams{
						Address:          node.IPAddress,
						AutonomousSystem: node.ASN,
						Vendor:           "contrail",
						RouterType:       "control-node",
						AdminDown:        false,
						Identifier:       node.IPAddress,
						HoldTime:         90,
						Port:             179,
						AddressFamilies: &contrailTypes.AddressFamilies{
							Family: []string{"route-target", "inet-vpn", "inet6-vpn", "e-vpn", "erm-vpn"},
						},
					}
					typedNode.SetBgpRouterParameters(bgpParameters)
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

// Delete deletes a ControlNode instance
func (c *ControlNode) Delete(nodeName string, contrailClient types.ApiClient) error {
	controlInfoLog.Println("Delete " + c.Hostname + " " + bgpRouterType)
	vncNodeList, err := contrailClient.List(bgpRouterType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(bgpRouterType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.BgpRouter)
		if typedNode.GetName() == nodeName {
			if !types.HasRequiredAnnotations(typedNode.GetAnnotations().KeyValuePair, c.Annotations) {
				controlInfoLog.Println(c.Hostname + " " + bgpRouterType + " does not have the required annotations.")
				controlInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + bgpRouterType)
				return nil
			}
			gscObjects := []*contrailTypes.GlobalSystemConfig{}
			gscObjectsList, err := contrailClient.List("global-system-config")
			if err != nil {
				return err
			}

			if len(gscObjectsList) == 0 {
				controlInfoLog.Println("no gscObject")
			}

			for _, gscObject := range gscObjectsList {
				obj, err := contrailClient.ReadListResult("global-system-config", &gscObject)
				if err != nil {
					return err
				}
				gscObjects = append(gscObjects, obj.(*contrailTypes.GlobalSystemConfig))
			}

			if len(gscObjects) > 0 {
				for _, gsc := range gscObjects {
					if err := gsc.DeleteBgpRouter(obj.GetUuid()); err != nil {
						return err
					}
					if err := contrailClient.Update(gsc); err != nil {
						return err
					}
				}
			}
			err = contrailClient.Delete(obj)
			if err != nil {
				return err
			}
		}

	}
	return nil
}


func ReconcileControlNodes(contrailClient types.ApiClient, nodeList []*ControlNode) error {
	var actionMap = make(map[string]string)
	nodeType := "bgp-router"
	vncNodes := []*ControlNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.BgpRouter)
		bgpRouterParamters := typedNode.GetBgpRouterParameters()
		if bgpRouterParamters.RouterType == "control-node" {
			node := &ControlNode{
				IPAddress: bgpRouterParamters.Address,
				Hostname:  typedNode.GetName(),
				ASN:       bgpRouterParamters.AutonomousSystem,
			}
			vncNodes = append(vncNodes, node)
		}
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
					if node.ASN != vncNode.ASN {
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
					controlInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					controlInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			node := &ControlNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
			controlInfoLog.Println("deleting node ", k)
		}
	}
	return nil
}
