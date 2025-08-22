/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/supabase"
	"github.com/train360-corp/projconf/internal/supabase/database"
	"net/http"
	"strings"
	"time"
)

func GetServerInterface() api.ServerInterface {
	return &RouteHandlers{}
}

// RouteHandlers implements api.ServerInterface (generated).
type RouteHandlers struct{}

func (s *RouteHandlers) GetV1ProjectsProjectIdVariables(c *gin.Context, projectId openapitypes.UUID) {
	sb := supabase.GetFromContext(c)
	id, _ := uuid.Parse(projectId.String())
	if variables, err := sb.GetVariables(&id); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Error:       "unable to get variables",
			Description: err.Error(),
		})
	} else {
		vars := make(api.Variables, len(*variables))
		var id uuid.UUID
		for i, e := range *variables {
			id, _ = uuid.Parse(e.Id)
			vars[i] = api.Variable{
				Description:   e.Description,
				GeneratorData: api.Variable_GeneratorData{},
				GeneratorType: api.GeneratorType(e.GeneratorType),
				Id:            id,
				Key:           e.Key,
				ProjectId:     projectId,
			}
		}
		c.JSON(http.StatusOK, vars)
	}
}

func (s *RouteHandlers) PostV1ProjectsProjectIdVariables(c *gin.Context, projectId openapitypes.UUID) {
	var req api.PostV1ProjectsProjectIdVariablesJSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	description := ""
	insert := database.PublicVariablesInsert{
		Description: &description,
		Id:          &id,
		Key:         req.Key,
		ProjectId:   projectId.String(),
	}

	typ, _ := req.Generator.Discriminator()
	switch api.GeneratorType(typ) {
	case api.GeneratorTypeRANDOM:
		r, _ := req.Generator.AsSecretGeneratorRandom()
		insert.GeneratorData = r.Data
		insert.GeneratorType = string(r.Type)
	case api.GeneratorTypeSTATIC:
		r, _ := req.Generator.AsSecretGeneratorStatic()
		insert.GeneratorData = r.Data
		insert.GeneratorType = string(r.Type)
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{
			Error:       "unable to create variable",
			Description: fmt.Sprintf("type \"%s\" is unhandled", typ),
		})
		return
	}

	if variable, err := sb.PostVariable(insert); err != nil {
		if strings.Index(err.Error(), "new row violates row-level security policy for table") != -1 {
			c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
				Error:       "unable to create variable",
				Description: "permission denied",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
				Error:       "unable to create variable",
				Description: err.Error(),
			})
		}
	} else {
		id, _ := uuid.Parse(variable.Id)
		c.JSON(http.StatusOK, api.ID{
			Id: id,
		})
	}
}

func (s *RouteHandlers) GetV1ProjectsProjectIdEnvironments(c *gin.Context, projectId openapitypes.UUID) {
	sb := supabase.GetFromContext(c)

	id, _ := uuid.Parse(projectId.String())
	if environments, err := sb.GetEnvironments(&id); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Error:       "unable to get environments",
			Description: err.Error(),
		})
	} else {
		envs := make(api.Environments, len(*environments))
		for i, e := range *environments {
			envs[i] = api.Environment{
				Id:      e.Id,
				Display: e.Display,
			}
		}
		c.JSON(http.StatusOK, envs)
	}
}

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

func (s *RouteHandlers) PostV1ProjectsProjectIdEnvironments(c *gin.Context, projectId openapitypes.UUID) {

	var req api.PostV1ProjectsProjectIdEnvironmentsJSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	createdAt := time.Now().Format(time.RFC3339)
	if environment, err := sb.PostEnvironment(database.PublicEnvironmentsInsert{
		Display:   &req.Name,
		ProjectId: projectId.String(),
		Id:        &id,
		CreatedAt: &createdAt,
	}); err != nil {
		if strings.Index(err.Error(), "new row violates row-level security policy for table") != -1 {
			c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
				Error:       "unable to create environment",
				Description: "permission denied",
			})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
				Error:       "unable to create environment",
				Description: err.Error(),
			})
		}
	} else {
		id, _ := uuid.Parse(environment.Id)
		c.JSON(http.StatusOK, api.ID{
			Id: id,
		})
	}
}

func (s *RouteHandlers) PostV1Projects(c *gin.Context) {

	var req api.PostV1ProjectsJSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	if project, err := sb.PostProject(database.PublicProjectsInsert{
		Display: &req.Name,
		Id:      &id,
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Error:       "unable to create project",
			Description: err.Error(),
		})
	} else {
		id, _ := uuid.Parse(project.Id)
		c.JSON(http.StatusOK, api.ID{
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
