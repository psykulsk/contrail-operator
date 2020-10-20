package databasenodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
)

var databaseNodesInfoLog *log.Logger

func init() {
	databaseNodesInfoLog = log.New(os.Stdout, "databasenodes: ", log.LstdFlags)
}

func ReconcileDatabaseNodes(contrailClient types.ApiClient, nodeList []*types.DatabaseNode) error {
	var actionMap = make(map[string]string)
	nodeType := "database-node"
	vncNodes := []*types.DatabaseNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return err
	}
	databaseNodesInfoLog.Printf("VncNodeList: %v\n", vncNodeList)
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return err
		}
		typedNode := obj.(*contrailTypes.DatabaseNode)

		node := &types.DatabaseNode{
			IPAddress: typedNode.GetDatabaseNodeIpAddress(),
			Hostname:  typedNode.GetName(),
		}
		vncNodes = append(vncNodes, node)
	}
	for _, node := range nodeList {
		actionMap[node.Hostname] = "create"
	}
	databaseNodesInfoLog.Printf("VncNodes: %v\n", vncNodes)

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
		databaseNodesInfoLog.Printf("actionMapValue: %v\n", v)
		switch v {
		case "update":
			for _, node := range nodeList {
				if node.Hostname == k {
					databaseNodesInfoLog.Println("updating node ", node.Hostname)
					err = node.Update(nodeList, k, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "create":
			for _, node := range nodeList {
				if node.Hostname == k {
					databaseNodesInfoLog.Println("creating node ", node.Hostname)
					err = node.Create(nodeList, node.Hostname, contrailClient)
					if err != nil {
						return err
					}
				}
			}
		case "delete":
			node := &types.DatabaseNode{}
			err = node.Delete(k, contrailClient)
			if err != nil {
				return err
			}
			databaseNodesInfoLog.Println("deleting node ", k)
		}
	}
	return nil
}
