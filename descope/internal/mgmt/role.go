package mgmt

import (
	"github.com/descope/go-sdk/descope"
	"github.com/descope/go-sdk/descope/api"
	"github.com/descope/go-sdk/descope/internal/utils"
)

type role struct {
	managementBase
}

func (r *role) Create(name, description string, permissionNames []string) error {
	if name == "" {
		return utils.NewInvalidArgumentError("name")
	}
	body := map[string]any{
		"name":            name,
		"description":     description,
		"permissionNames": permissionNames,
	}
	_, err := r.client.DoPostRequest(api.Routes.ManagementRoleCreate(), body, nil, r.conf.ManagementKey)
	return err
}

func (r *role) Update(name, newName, description string, permissionNames []string) error {
	if name == "" {
		return utils.NewInvalidArgumentError("name")
	}
	if newName == "" {
		return utils.NewInvalidArgumentError("newName")
	}
	body := map[string]any{
		"name":            name,
		"newName":         newName,
		"description":     description,
		"permissionNames": permissionNames,
	}
	_, err := r.client.DoPostRequest(api.Routes.ManagementRoleUpdate(), body, nil, r.conf.ManagementKey)
	return err
}

func (r *role) Delete(name string) error {
	if name == "" {
		return utils.NewInvalidArgumentError("name")
	}
	body := map[string]any{"name": name}
	_, err := r.client.DoPostRequest(api.Routes.ManagementRoleDelete(), body, nil, r.conf.ManagementKey)
	return err
}

func (r *role) LoadAll() ([]*descope.Role, error) {
	res, err := r.client.DoGetRequest(api.Routes.ManagementRoleLoadAll(), nil, r.conf.ManagementKey)
	if err != nil {
		return nil, err
	}
	return unmarshalRolesLoadAllResponse(res)
}

func unmarshalRolesLoadAllResponse(res *api.HTTPResponse) ([]*descope.Role, error) {
	pres := struct {
		Roles []*descope.Role
	}{}
	err := utils.Unmarshal([]byte(res.BodyStr), &pres)
	if err != nil {
		return nil, err
	}
	return pres.Roles, nil
}
