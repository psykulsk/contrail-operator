package types

import (
	contrail "github.com/Juniper/contrail-go-api"

	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
)

const (
	managedByAnnotationKey = "managed_by"
)

// Nodes struct defines all Contrail node types
type Nodes struct {
	ControlNodes   []*ControlNode             `yaml:"controlNodes,omitempty"`
	BgpRouters     []*contrailTypes.BgpRouter `yaml:"bgpRouters,omitempty"`
	AnalyticsNodes []*AnalyticsNode           `yaml:"analyticsNodes,omitempty"`
	VrouterNodes   []*VrouterNode             `yaml:"vrouterNodes,omitempty"`
	ConfigNodes    []*ConfigNode              `yaml:"configNodes,omitempty"`
	DatabaseNodes  []*DatabaseNode            `yaml:"databaseNodes,omitempty"`
}

// ApiClient interface extends contrail.ApiClient by a missing ReadListResult
// to enable passing ApiClient interface instead of the struct to ease
// mocking in unit test
type ApiClient interface {
	contrail.ApiClient
	ReadListResult(string, *contrail.ListResult) (contrail.IObject, error)
}

func IsManagedByMe(annotations []contrailTypes.KeyValuePair, provisionManagerName string) bool {
	managedByMe := false
	for _, annotation := range annotations {
		if annotation.Key == managedByAnnotationKey && annotation.Value == provisionManagerName {
			managedByMe = true
		}
	}
	return managedByMe
}

func GetManagedByMeAnnotation(provisionManagerName string) *contrailTypes.KeyValuePairs {
	return &contrailTypes.KeyValuePairs{
		KeyValuePair: []contrailTypes.KeyValuePair{
			{Key: managedByAnnotationKey, Value: provisionManagerName},
		},
	}
}
