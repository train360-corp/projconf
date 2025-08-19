/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"github.com/gin-gonic/gin"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/supabase"
	"github.com/train360-corp/projconf/internal/supabase/database"
	"github.com/train360-corp/projconf/internal/utils/postgres"
	"net/http"
	"time"
)

// RouteHandlers implements api.ServerInterface (generated).
type RouteHandlers struct{}

func (s *RouteHandlers) GetV1Projects(c *gin.Context) {
	sb := supabase.GetFromContext(c)
	if projects, err := sb.GetProjects(); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Error:       "unable to get projects",
			Description: err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, projects)
	}
}

func (s *RouteHandlers) PostV1AdminProjectsProjectIdEnvironments(c *gin.Context, projectId openapitypes.UUID) {

	var req struct {
		Name string `json:"name"`
	}
	c.BindJSON(&req)

	var id string
	err := postgres.Insert(c, "public.environments", database.PublicEnvironmentsInsert{
		Display:   &req.Name,
		ProjectId: projectId.String(),
	}, "id", &id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Error:       "unable to create environment",
			Description: err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, struct {
			Id string `json:"id"`
		}{
			Id: id,
		})
	}

}

func (s *RouteHandlers) PostV1AdminProjects(c *gin.Context) {

	var req struct {
		Name string `json:"name"`
	}
	c.BindJSON(&req)

	var id string
	err := postgres.Insert(c, "public.projects", database.PublicProjectsInsert{
		Display: &req.Name,
	}, "id", &id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Error:       "unable to create project",
			Description: err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, struct {
			Id string `json:"id"`
		}{
			Id: id,
		})
	}
}

func (s *RouteHandlers) GetV1AdminHealth(c *gin.Context) {
	c.JSON(http.StatusOK, struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Version   string `json:"version"`
	}{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   config.Version,
	})
}

func (s *RouteHandlers) GetV1ClientsSelf(c *gin.Context) {

	sb := supabase.GetFromContext(c)
	if client, clientErr := sb.GetSelf(); clientErr != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Description: clientErr.Error(),
			Error:       "unable to get client",
		})
	} else {
		c.JSON(http.StatusOK, client)
	}
}

type SecretKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *RouteHandlers) GetV1ProjectsProjectIdEnvironmentsEnvironmentIdSecrets(
	c *gin.Context,
	projectId openapitypes.UUID,
	environmentId openapitypes.UUID,
) {
	sb := supabase.GetFromContext(c)

	secrets, err := sb.GetSecrets(projectId.String(), environmentId.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: err.Error(),
			Error:       "unable to get secrets",
		})
		return
	}

	if len(secrets) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: "the array of secrets returned was empty",
			Error:       "empty secrets",
		})
		return
	}

	resp := make([]SecretKV, 0, len(secrets))
	for _, s := range secrets {
		resp = append(resp, SecretKV{
			Key:   s.Variables.Key,
			Value: s.Value,
		})
	}

	c.JSON(http.StatusOK, resp)
}
