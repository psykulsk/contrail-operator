package manager

import (
	core "k8s.io/api/core/v1"

	contrail "github.com/Juniper/contrail-operator/pkg/apis/contrail/v1alpha1"
	"github.com/Juniper/contrail-operator/pkg/k8s"
	"github.com/Juniper/contrail-operator/pkg/randomstring"
)

type keystoneSecret struct {
	sc *k8s.Secret
}

func (s *keystoneSecret) FillSecret(sc *core.Secret) error {
	if sc.Data != nil {
		return nil
	}

	pass := randomstring.RandString{10}.Generate()

	sc.StringData = map[string]string{
		"password": pass,
	}
	return nil
}

func (r *ReconcileManager) keystoneSecret(secretName, ownerType string, manager *contrail.Manager) *keystoneSecret {
	return &keystoneSecret{
		sc: r.kubernetes.Secret(secretName, ownerType, manager),
	}
}

func (s *keystoneSecret) ensureAdminPassSecretExist() error {
	return s.sc.EnsureExists(s)
}
