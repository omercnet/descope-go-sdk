package mgmt

import (
	"github.com/descope/go-sdk/descope"
	"github.com/descope/go-sdk/descope/api"
	"github.com/descope/go-sdk/descope/internal/utils"
)

type group struct {
	managementBase
}

func (r *group) LoadAllGroups(tenantID string) ([]*descope.Group, error) {
	if tenantID == "" {
		return nil, utils.NewInvalidArgumentError("tenantID")
	}
	body := map[string]any{
		"tenantId": tenantID,
	}
	res, err := r.client.DoPostRequest(api.Routes.ManagementGroupLoadAllGroups(), body, nil, r.conf.ManagementKey)
	if err != nil {
		return nil, err
	}
	return unmarshalGroupsResponse(res)
}

func (r *group) LoadAllGroupsForMembers(tenantID string, userIDs, loginIDs []string) ([]*descope.Group, error) {
	if tenantID == "" {
		return nil, utils.NewInvalidArgumentError("tenantID")
	}
	if len(userIDs) == 0 && len(loginIDs) == 0 {
		return nil, utils.NewInvalidArgumentError("userIDs and loginIDs")
	}
	body := map[string]any{
		"tenantId": tenantID,
		"loginIds": loginIDs,
		"userIds":  userIDs,
	}
	res, err := r.client.DoPostRequest(api.Routes.ManagementGroupLoadAllGroupsForMember(), body, nil, r.conf.ManagementKey)
	if err != nil {
		return nil, err
	}
	return unmarshalGroupsResponse(res)
}

func (r *group) LoadAllGroupMembers(tenantID, groupID string) ([]*descope.Group, error) {
	if tenantID == "" {
		return nil, utils.NewInvalidArgumentError("tenantID")
	}
	if groupID == "" {
		return nil, utils.NewInvalidArgumentError("groupID")
	}
	body := map[string]any{
		"tenantId": tenantID,
		"groupId":  groupID,
	}
	res, err := r.client.DoPostRequest(api.Routes.ManagementGroupLoadAllGroupMembers(), body, nil, r.conf.ManagementKey)
	if err != nil {
		return nil, err
	}
	return unmarshalGroupsResponse(res)
}

func unmarshalGroupsResponse(res *api.HTTPResponse) ([]*descope.Group, error) {
	var groups []*descope.Group
	err := utils.Unmarshal([]byte(res.BodyStr), &groups)
	if err != nil {
		// notest
		return nil, err
	}
	return groups, nil
}
