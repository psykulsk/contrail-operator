package controlnodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
)

var controlNodesInfoLog *log.Logger

func init() {
	controlNodesInfoLog = log.New(os.Stdout, "controlnodes: ", log.LstdFlags)
}

func ReconcileControlNodes(contrailClient types.ApiClient, nodeList []*types.ControlNode) error {
	var actionMap = make(map[string]string)
	nodeType := "bgp-router"
	vncNodes := []*types.ControlNode{}
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
			node := &types.ControlNode{
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
					controlNodesInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					controlNodesInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			node := &types.ControlNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
			controlNodesInfoLog.Println("deleting node ", k)
		}
	}
	return nil
}
