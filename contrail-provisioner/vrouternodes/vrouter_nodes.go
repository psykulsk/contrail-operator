package vrouternodes

import (
	"log"
	"os"

	"github.com/Juniper/contrail-go-api"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/contrailclient"
)

// VrouterNode struct defines Contrail Vrouter node
type VrouterNode struct {
	IPAddress   string            `yaml:"ipAddress,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

const (
	ipFabricNetworkFQName       = "default-domain:default-project:ip-fabric"
	vhost0VMIName               = "vhost0"
	virtualMachineInterfaceType = "virtual-machine-interface"
	nodeType                    = "virtual-router"
)

type Action int

const (
	updateAction Action = iota
	createAction
	deleteAction
	noopAction
)

type NodeWithAction struct {
	node   *VrouterNode
	action Action
}

var vrouterInfoLog *log.Logger

func init() {
	vrouterInfoLog = log.New(os.Stdout, "vrouternodes: ", log.LstdFlags)
}

// Create creates a VirtualRouter instance
func (c *VrouterNode) Create(contrailClient contrailclient.ApiClient) error {
	vrouterInfoLog.Println("Creating " + c.Hostname + " " + nodeType)
	gscObjects := []*contrailTypes.GlobalSystemConfig{}
	gscObjectsList, err := contrailClient.List("global-system-config")
	if err != nil {
		return err
	}

	if len(gscObjectsList) == 0 {
		vrouterInfoLog.Println("no gscObject")
	}

	for _, gscObject := range gscObjectsList {
		obj, err := contrailClient.ReadListResult("global-system-config", &gscObject)
		if err != nil {
			return err
		}
		gscObjects = append(gscObjects, obj.(*contrailTypes.GlobalSystemConfig))
	}
	for _, gsc := range gscObjects {
		virtualRouter := &contrailTypes.VirtualRouter{}
		virtualRouter.SetVirtualRouterIpAddress(c.IPAddress)
		virtualRouter.SetParent(gsc)
		virtualRouter.SetName(c.Hostname)
		annotations := &contrailTypes.KeyValuePairs{
			KeyValuePair: contrailclient.ConvertMapToContrailKeyValuePairs(c.Annotations),
		}
		virtualRouter.SetAnnotations(annotations)
		if err := contrailClient.Create(virtualRouter); err != nil {
			return err
		}
		return nil
	}
	return nil
}

// Update updates a VirtualRouter instance
func (c *VrouterNode) Update(contrailClient contrailclient.ApiClient) error {
	vrouterInfoLog.Println("Updating " + c.Hostname + " " + nodeType)
	obj, err := contrailclient.GetContrailObjectByName(contrailClient, nodeType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !contrailclient.HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + nodeType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + nodeType)
		return nil
	}
	virtualRouter.SetVirtualRouterIpAddress(c.IPAddress)
	if err := contrailClient.Update(virtualRouter); err != nil {
		return err
	}
	return nil
}

// Delete deletes a VirtualRouter instance and it's vhost0 VirtualMachineInterfaces
func (c *VrouterNode) Delete(contrailClient contrailclient.ApiClient) error {
	vrouterInfoLog.Println("Deleting " + c.Hostname + " " + nodeType)
	obj, err := contrailclient.GetContrailObjectByName(contrailClient, nodeType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !contrailclient.HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + nodeType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + nodeType)
		return nil
	}
	deleteVMIs(virtualRouter, contrailClient)
	vrouterInfoLog.Println("Deleting VirtualRouter ", obj.GetName())
	if err = contrailClient.Delete(obj); err != nil {
		return err
	}
	return nil
}

func (c *VrouterNode) EnsureVMIVhost0Interface(contrailClient contrailclient.ApiClient) error {
	vrouterInfoLog.Printf("Ensuring %v %v has the vhost0 virtual-machine interface assigned", c.Hostname, nodeType)
	obj, err := contrailclient.GetContrailObjectByName(contrailClient, nodeType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !contrailclient.HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + nodeType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping virtual-machine-interface modifications for " + c.Hostname + " " + nodeType)
		return nil
	}
	return ensureVMIVhost0Interface(contrailClient, virtualRouter)
}

// EnsureVMIVhost0Interface checks whether the VirtualRouter
// has a "vhost0" VirtualMachineInterface assigned to it and creates
// one if neccessary.
func ensureVMIVhost0Interface(contrailClient contrailclient.ApiClient, virtualRouter *contrailTypes.VirtualRouter) error {
	vhost0VMIPresent, err := vhost0VMIPresent(virtualRouter, contrailClient)
	if err != nil {
		return err
	}
	if vhost0VMIPresent {
		vrouterInfoLog.Printf("vhost0 virtual-machine-interface already exists for %v %v\n", virtualRouter.GetName(), nodeType)
		return nil
	}
	return createVhost0VMI(virtualRouter, contrailClient)
}

func vhost0VMIPresent(virtualRouter *contrailTypes.VirtualRouter, contrailClient contrailclient.ApiClient) (bool, error) {
	vmiList, err := virtualRouter.GetVirtualMachineInterfaces()
	if err != nil {
		return false, err
	}
	for _, vmiRef := range vmiList {
		vmiObj, err := contrailClient.FindByUuid(virtualMachineInterfaceType, vmiRef.Uuid)
		if err != nil {
			return false, err
		}
		if vmiObj.GetName() == vhost0VMIName {
			return true, nil
		}
	}
	return false, nil
}

func createVhost0VMI(virtualRouter *contrailTypes.VirtualRouter, contrailClient contrailclient.ApiClient) error {
	network, err := contrailTypes.VirtualNetworkByName(contrailClient, ipFabricNetworkFQName)
	if err != nil {
		return err
	}
	vncVMI := &contrailTypes.VirtualMachineInterface{}
	vrouterInfoLog.Println("Creating vhost0 virtual-machine-interface for VirtualRouter: ", virtualRouter.GetName())
	vncVMI.SetParent(virtualRouter)
	vncVMI.SetVirtualNetworkList([]contrail.ReferencePair{{Object: network}})
	vncVMI.SetVirtualMachineInterfaceDisablePolicy(false)
	vncVMI.SetName(vhost0VMIName)
	if err = contrailClient.Create(vncVMI); err != nil {
		return err
	}
	return nil
}

func deleteVMIs(virtualRouter *contrailTypes.VirtualRouter, contrailClient contrailclient.ApiClient) error {
	vmiList, err := virtualRouter.GetVirtualMachineInterfaces()
	if err != nil {
		return err
	}
	for _, vmiRef := range vmiList {
		vrouterInfoLog.Println("Deleting virtual-machine-interface ", vmiRef.Uuid)
		if err = contrailClient.DeleteByUuid(virtualMachineInterfaceType, vmiRef.Uuid); err != nil {
			return err
		}
	}
	return nil
}

// ReconcileVrouterNodes creates, deletes or updates VirtualRouter and
// VirtualMachineInterface objects in Contrail Api Server based on the
// list of requiredNodes and current objects in the Api Server
func ReconcileVrouterNodes(contrailClient contrailclient.ApiClient, requiredNodes []*VrouterNode) error {
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

func getVrouterNodesInApiServer(contrailClient contrailclient.ApiClient) ([]*VrouterNode, error) {
	nodesInApiServer := []*VrouterNode{}
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

		node := &VrouterNode{
			IPAddress: typedNode.GetVirtualRouterIpAddress(),
			Hostname:  typedNode.GetName(),
		}
		nodesInApiServer = append(nodesInApiServer, node)
	}

	return nodesInApiServer, nil
}

func createVrouterNodesActionMap(nodesInApiServer []*VrouterNode, requiredNodes []*VrouterNode) map[string]NodeWithAction {
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

func executeActionMap(actionMap map[string]NodeWithAction, contrailClient contrailclient.ApiClient) error {
	for _, nodeWithAction := range actionMap {
		var err error
		switch nodeWithAction.action {
		case updateAction:
			vrouterInfoLog.Println("updating vrouter node ", nodeWithAction.node.Hostname)
			err = nodeWithAction.node.Update(contrailClient)
		case createAction:
			vrouterInfoLog.Println("creating vrouter node ", nodeWithAction.node.Hostname)
			err = nodeWithAction.node.Create(contrailClient)
		case deleteAction:
			vrouterInfoLog.Println("deleting vrouter node ", nodeWithAction.node.Hostname)
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
