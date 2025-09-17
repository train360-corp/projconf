/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/train360-corp/projconf/go/pkg"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/server/state"
	"github.com/train360-corp/projconf/go/pkg/supabase"
	database2 "github.com/train360-corp/projconf/go/pkg/supabase/database"
	"net/http"
	"strings"
	"time"
)

func GetServerInterface() api.ServerInterface {
	return &RouteHandlers{}
}

// RouteHandlers implements api.ServerInterface (generated).
type RouteHandlers struct{}

func (s *RouteHandlers) GetStatusV1(c *gin.Context) {
	state := state.Get()
	c.JSON(http.StatusOK, api.Status{
		Server: struct {
			IsReady bool   `json:"is_ready"`
			Version string `json:"version"`
		}{
			IsReady: state.IsAlive(),
			Version: pkg.Version,
		},
		Services: struct {
			Postgres  bool `json:"postgres"`
			Postgrest bool `json:"postgrest"`
		}{
			Postgres:  state.IsPostgresAlive(),
			Postgrest: state.IsPostgrestAlive(),
		},
	})
}

func (s *RouteHandlers) GetStatusReadyV1(c *gin.Context) {
	state := state.Get()
	if state.IsAlive() {
		c.JSON(http.StatusOK, api.Ready{
			Msg: "ready",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, api.ServiceUnavailable{
			Error:       "Unavailable",
			Description: "one or more required services are not available",
		})
	}
}

func (s *RouteHandlers) GetClientsV1(c *gin.Context, environmentId openapitypes.UUID) {
	sb := supabase.GetFromContext(c)
	id, _ := uuid.Parse(environmentId.String())
	if clients, err := sb.GetClients(&id); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Error:       "unable to get clients",
			Description: err.Error(),
		})
	} else {
		envs := make([]api.ClientRepresentation, len(*clients))
		for i, c := range *clients {
			cId, _ := uuid.Parse(c.Id)
			eId, _ := uuid.Parse(c.EnvironmentId)
			envs[i] = api.ClientRepresentation{
				Id:            cId,
				Display:       c.Display,
				CreatedAt:     c.CreatedAt,
				EnvironmentId: eId,
			}
		}
		c.JSON(http.StatusOK, envs)
	}
}

func (s *RouteHandlers) CreateClientV1(c *gin.Context, environmentId openapitypes.UUID) {

	var req api.CreateClientV1JSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	resp, err := sb.CreateClient(database2.PublicRpcCreateClientAndSecretRequest{
		Display:       req.Name,
		EnvironmentId: environmentId.String(),
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: err.Error(),
			Error:       "unable to create client",
		})
		return
	}

	id, _ := uuid.Parse(resp.ClientId)
	secretId, _ := uuid.Parse(resp.SecretId)
	c.JSON(http.StatusCreated, api.CreateClientResponse{
		Id: id,
		Secret: struct {
			Id  openapitypes.UUID `json:"id"`
			Key string            `json:"key"`
		}{
			Id:  secretId,
			Key: resp.Secret,
		},
	})
	return
}

func (s *RouteHandlers) GetClientSecretsV1(c *gin.Context) {
	sb := supabase.GetFromContext(c)

	client, err := sb.GetSelf()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: err.Error(),
			Error:       "unable to get secrets",
		})
		return
	}

	secrets, err := sb.GetSecrets(client.EnvironmentId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: err.Error(),
			Error:       "unable to get secrets",
		})
		return
	}
	c.JSON(http.StatusOK, secrets)
}

func (s *RouteHandlers) GetVariablesV1(c *gin.Context, projectId openapitypes.UUID) {
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

func (s *RouteHandlers) CreateVariableV1(c *gin.Context, projectId openapitypes.UUID) {
	var req api.CreateVariableV1JSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	description := ""
	insert := database2.PublicVariablesInsert{
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
		c.JSON(http.StatusCreated, api.ID{
			Id: id,
		})
	}
}

func (s *RouteHandlers) GetEnvironmentsV1(c *gin.Context, projectId openapitypes.UUID) {
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

func (s *RouteHandlers) CreateEnvironmentV1(c *gin.Context, projectId openapitypes.UUID) {
	var req api.CreateEnvironmentV1JSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	createdAt := time.Now().Format(time.RFC3339)
	if environment, err := sb.PostEnvironment(database2.PublicEnvironmentsInsert{
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
		c.JSON(http.StatusCreated, api.ID{
			Id: id,
		})
	}
}

func (s *RouteHandlers) GetProjectsV1(c *gin.Context) {
	sb := supabase.GetFromContext(c)
	if projects, err := sb.GetProjects(); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Error:       "unable to get projects",
			Description: err.Error(),
		})
	} else {
		var resp api.Projects
		for _, p := range *projects {
			id, _ := uuid.Parse(p.Id)
			resp = append(resp, api.Project{
				Id:      id,
				Display: p.Display,
			})
		}
		c.JSON(http.StatusOK, resp)
	}
}

func (s *RouteHandlers) CreateProjectV1(c *gin.Context) {
	var req api.CreateProjectV1JSONRequestBody
	c.BindJSON(&req)

	sb := supabase.GetFromContext(c)
	id := uuid.New().String()
	if project, err := sb.PostProject(database2.PublicProjectsInsert{
		Display: &req.Name,
		Id:      &id,
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Error:       "unable to create project",
			Description: err.Error(),
		})
	} else {
		id, _ := uuid.Parse(project.Id)
		c.JSON(http.StatusCreated, api.ID{
			Id: id,
		})
	}
}

// GetV1ClientsSelf an authenticated client retrieving itself
func (s *RouteHandlers) GetV1ClientsSelf(c *gin.Context) {
	sb := supabase.GetFromContext(c)
	if client, clientErr := sb.GetSelf(); clientErr != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, api.Error{
			Description: clientErr.Error(),
			Error:       "unable to get client",
		})
	} else {
		envId, _ := uuid.Parse(client.EnvironmentId)
		id, _ := uuid.Parse(client.Id)
		c.JSON(http.StatusOK, api.ClientRepresentation{
			CreatedAt:     client.CreatedAt,
			Display:       client.Display,
			EnvironmentId: envId,
			Id:            id,
		})
	}
}

func (s *RouteHandlers) GetEnvironmentSecretsV1(
	c *gin.Context,
	environmentId openapitypes.UUID,
) {
	sb := supabase.GetFromContext(c)
	secrets, err := sb.GetSecrets(environmentId.String())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.Error{
			Description: err.Error(),
			Error:       "unable to get secrets",
		})
		return
	}
	c.JSON(http.StatusOK, secrets)
}
