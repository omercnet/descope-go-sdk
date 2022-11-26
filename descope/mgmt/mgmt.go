package mgmt

import (
	"github.com/descope/go-sdk/descope/api"
	"github.com/descope/go-sdk/descope/logger"
	"github.com/descope/go-sdk/descope/utils"
)

type ManagementParams struct {
	ProjectID     string
	ManagementKey string
}

type managementBase struct {
	client *api.Client
	conf   *ManagementParams
}

type managementService struct {
	managementBase

	tenant Tenant
	user   User
	sso    SSO
}

func NewManagement(conf ManagementParams, c *api.Client) *managementService {
	base := managementBase{conf: &conf, client: c}
	service := &managementService{managementBase: base}
	service.tenant = &tenant{managementBase: base}
	service.user = &user{managementBase: base}
	service.sso = &sso{managementBase: base}
	return service
}

func (mgmt *managementService) Tenant() Tenant {
	mgmt.ensureManagementKey()
	return mgmt.tenant
}

func (mgmt *managementService) User() User {
	mgmt.ensureManagementKey()
	return mgmt.user
}

func (mgmt *managementService) SSO() SSO {
	mgmt.ensureManagementKey()
	return mgmt.sso
}

func (mgmt *managementService) ensureManagementKey() {
	if mgmt.conf.ManagementKey == "" {
		logger.LogInfo("management key is missing, make sure to add it in the Config struct or the environment variable \"%s\"", utils.EnvironmentVariableManagementKey) // notest
	}
}