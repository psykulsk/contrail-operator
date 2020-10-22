package vrouternodes

import (
	"log"
	"os"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/types"
)

const nodeType string = "virtual-router"

type Action int

const (
	updateAction Action = iota
	createAction
	deleteAction
	noopAction
)

type NodeWithAction struct {
	node   *types.VrouterNode
	action Action
}

var vrouterNodesInfoLog *log.Logger

func init() {
	vrouterNodesInfoLog = log.New(os.Stdout, "vrouternodes: ", log.LstdFlags)
}

// ReconcileVrouterNodes creates, deletes or updates VirtualRouter and
// VirtualMachineInterface objects in Contrail Api Server based on the
// list of requiredNodes and current objects in the Api Server
func ReconcileVrouterNodes(contrailClient types.ApiClient, requiredNodes []*types.VrouterNode) error {
	nodesInApiServer, err := getVrouterNodesInApiServer(contrailClient)
	if err != nil {
		return err
	}
	actionMap := createVrouterNodesActionMap(nodesInApiServer, requiredNodes)
	if err = executeActionMap(actionMap, contrailClient); err != nil {
		return err
	}
	return nil
}

func getVrouterNodesInApiServer(contrailClient types.ApiClient) ([]*types.VrouterNode, error) {
	nodesInApiServer := []*types.VrouterNode{}
	vncNodeList, err := contrailClient.List(nodeType)
	if err != nil {
		return nodesInApiServer, err
	}
	for _, vncNode := range vncNodeList {
		obj, err := contrailClient.ReadListResult(nodeType, &vncNode)
		if err != nil {
			return nodesInApiServer, err
		}
		typedNode := obj.(*contrailTypes.VirtualRouter)

		node := &types.VrouterNode{
			IPAddress: typedNode.GetVirtualRouterIpAddress(),
			Hostname:  typedNode.GetName(),
		}
		nodesInApiServer = append(nodesInApiServer, node)
	}

	return nodesInApiServer, nil
}

func createVrouterNodesActionMap(nodesInApiServer []*types.VrouterNode, requiredNodes []*types.VrouterNode) map[string]NodeWithAction {
	var actionMap = make(map[string]NodeWithAction)
	for _, requiredNode := range requiredNodes {
		actionMap[requiredNode.Hostname] = NodeWithAction{node: requiredNode, action: createAction}
	}
	for _, nodeInApiServer := range nodesInApiServer {
		if requiredNodeWithAction, ok := actionMap[nodeInApiServer.Hostname]; ok {
			requiredAction := noopAction
			if requiredNodeWithAction.node.IPAddress != nodeInApiServer.IPAddress {
				requiredAction = updateAction
			}
			actionMap[nodeInApiServer.Hostname] = NodeWithAction{
				node:   requiredNodeWithAction.node,
				action: requiredAction,
			}
		} else {
			actionMap[nodeInApiServer.Hostname] = NodeWithAction{
				node:   nodeInApiServer,
				action: deleteAction,
			}
		}
	}
	return actionMap
}

func executeActionMap(actionMap map[string]NodeWithAction, contrailClient types.ApiClient) error {
	for _, nodeWithAction := range actionMap {
		var err error
		switch nodeWithAction.action {
		case updateAction:
			vrouterNodesInfoLog.Println("updating vrouter node ", nodeWithAction.node.Hostname)
			err = nodeWithAction.node.Update(contrailClient)
		case createAction:
			vrouterNodesInfoLog.Println("creating vrouter node ", nodeWithAction.node.Hostname)
			err = nodeWithAction.node.Create(contrailClient)
		case deleteAction:
			vrouterNodesInfoLog.Println("deleting vrouter node ", nodeWithAction.node.Hostname)
			err = nodeWithAction.node.Delete(contrailClient)
		}
		if err != nil {
			return err
		}
		// Ensure that all non-deleted vrouter nodes have their
		// respective vhost0 virtual-machine-interfaces created
		if nodeWithAction.action != deleteAction {
			if err := nodeWithAction.node.EnsureVMIVhost0Interface(contrailClient); err != nil {
				return err
			}
		}
	}
	return nil
}
