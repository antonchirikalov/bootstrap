package authorization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/rs/zerolog"
)

// access represents a response of access endpoint
type Access struct {
	SuperUser        bool            `json:"SuperUser"`
	Admin            *AdminType      `json:"Admin,omitempty"`
	VOAdmin          *AdminType      `json:"VOAdmin,omitempty"`
	VONoChildAdmin   *AdminType      `json:"VONoChildAdmin,omitempty"`
	FSAdmin          *FsAdminType    `json:"FSAdmin,omitempty"`
	FSVOAdmin        *FsAdminType    `json:"FSVOAdmin,omitempty"`
	Teacher          *TeacherType    `json:"Teacher,omitempty"`
	CoTeacher        *TeacherType    `json:"CoTeacher,omitempty"`
	AssistantTeacher *TeacherType    `json:"AssistantTeacher,omitempty"`
	TeamMember       *TeamMemberType `json:"TeamMember,omitempty"`
}

type AdminType struct {
	Ent []int64 `json:"ent,omitempty"`
}

type FsAdminType struct {
	AdminType
	FundSrc []int64 `json:"fundSrc,omitempty"`
}

type TeacherType struct {
	Cls []int64 `json:"cls,omitempty"`
}

type TeamMemberType struct {
	Kid []int64 `json:"kid,omitempty"`
}

// LoadAccess loads access data from authorization service for given userID
func LoadAccess(authorizationServiceURL, userID, token string, logger *zerolog.Logger) (*Access, error) {
	serverURL, err := url.Parse(authorizationServiceURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse authorization service url: %v", err)
	}
	serverURL.Path = path.Join(serverURL.Path, "access", userID)

	req, err := http.NewRequest("GET", serverURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a request to authorization service: %v", err)
	}
	req.Header.Set("iam", token)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error().Err(err).Msg("unable to close body of access request")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to retrieve access data '%s', status %d", resp.Status, resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Message string
		Data    Access
		Status  string
	}
	err = json.Unmarshal(b, &payload)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal result from authorization service: %v", err)
	}
	if payload.Status != "success" {
		return nil, errors.New(payload.Message)
	}
	return &payload.Data, nil
}
