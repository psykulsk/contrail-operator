package types

import (
	"log"
	"os"

	"github.com/Juniper/contrail-go-api"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
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
	virtualRouterType           = "virtual-router"
)

var vrouterInfoLog *log.Logger

func init() {
	vrouterInfoLog = log.New(os.Stdout, "vrouter: ", log.LstdFlags)
}

// Create creates a VirtualRouter instance
func (c *VrouterNode) Create(contrailClient ApiClient) error {
	vrouterInfoLog.Println("Creating " + c.Hostname + " " + virtualRouterType)
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
			KeyValuePair: ConvertMapToContrailKeyValuePairs(c.Annotations),
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
func (c *VrouterNode) Update(contrailClient ApiClient) error {
	vrouterInfoLog.Println("Updating " + c.Hostname + " " + virtualRouterType)
	obj, err := GetContrailObjectByName(contrailClient, virtualRouterType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + virtualRouterType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping Update operation of " + c.Hostname + " " + virtualRouterType)
		return nil
	}
	virtualRouter.SetVirtualRouterIpAddress(c.IPAddress)
	if err := contrailClient.Update(virtualRouter); err != nil {
		return err
	}
	return nil
}

// Delete deletes a VirtualRouter instance and it's vhost0 VirtualMachineInterfaces
func (c *VrouterNode) Delete(contrailClient ApiClient) error {
	vrouterInfoLog.Println("Deleting " + c.Hostname + " " + virtualRouterType)
	obj, err := GetContrailObjectByName(contrailClient, virtualRouterType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + virtualRouterType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping Delete operation of " + c.Hostname + " " + virtualRouterType)
		return nil
	}
	deleteVMIs(virtualRouter, contrailClient)
	vrouterInfoLog.Println("Deleting VirtualRouter ", obj.GetName())
	if err = contrailClient.Delete(obj); err != nil {
		return err
	}
	return nil
}

func (c *VrouterNode) EnsureVMIVhost0Interface(contrailClient ApiClient) error {
	vrouterInfoLog.Println("Ensuring " + c.Hostname + " " + virtualRouterType + " has the vhost0 virtual-machine interface assigned")
	obj, err := GetContrailObjectByName(contrailClient, virtualRouterType, c.Hostname)
	if err != nil {
		return err
	}
	virtualRouter := obj.(*contrailTypes.VirtualRouter)
	if !HasRequiredAnnotations(virtualRouter.GetAnnotations().KeyValuePair, c.Annotations) {
		vrouterInfoLog.Println(c.Hostname + " " + virtualRouterType + " does not have the required annotations.")
		vrouterInfoLog.Println("Skipping virtual-machine-interface modifications for " + c.Hostname + " " + virtualRouterType)
		return nil
	}
	return ensureVMIVhost0Interface(contrailClient, virtualRouter)
}

// EnsureVMIVhost0Interface checks whether the VirtualRouter
// has a "vhost0" VirtualMachineInterface assigned to it and creates
// one if neccessary.
func ensureVMIVhost0Interface(contrailClient ApiClient, virtualRouter *contrailTypes.VirtualRouter) error {
	vhost0VMIPresent, err := vhost0VMIPresent(virtualRouter, contrailClient)
	if err != nil {
		return err
	}
	if !vhost0VMIPresent {
		if err = createVhost0VMI(virtualRouter, contrailClient); err != nil {
			return err
		}
	}
	return nil
}

func vhost0VMIPresent(virtualRouter *contrailTypes.VirtualRouter, contrailClient ApiClient) (bool, error) {
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

func createVhost0VMI(virtualRouter *contrailTypes.VirtualRouter, contrailClient ApiClient) error {
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

func deleteVMIs(virtualRouter *contrailTypes.VirtualRouter, contrailClient ApiClient) error {
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
