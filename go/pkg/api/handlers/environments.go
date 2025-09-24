/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/train360-corp/projconf/go/internal/utils"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/postgrest"
	"github.com/train360-corp/projconf/go/pkg/server/state"
	"net/http"
)

func (r RouteHandlers) DeleteEnvironmentV1(c *gin.Context, id api.ID) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.DeleteEnvironmentsWithResponse(context.Background(), &postgrest.DeleteEnvironmentsParams{Id: equals(id)}); err != nil {
		state.Get().GetLogger().Debugf("[%s] request failed: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: "a pre-flight error occurred while processing the upstream request",
		})
	} else if response.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: fmt.Sprintf("error %d", response.StatusCode()),
		})
	} else {
		c.JSON(http.StatusOK, success)
	}
}

func (r RouteHandlers) GetEnvironmentV1(c *gin.Context, id api.ID) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.GetEnvironmentsWithResponse(context.Background(), &postgrest.GetEnvironmentsParams{Id: equals(id)}); err != nil {
		state.Get().GetLogger().Debugf("[%s] request failed: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: "a pre-flight error occurred while processing the upstream request",
		})
	} else if response.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: fmt.Sprintf("error %d", response.StatusCode()),
		})
	} else if environments, err := parse[[]postgrest.Environments](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else if len(*environments) == 0 {
		c.JSON(http.StatusNotFound, &api.Error{
			Error:       "not found",
			Description: fmt.Sprintf("an environment with id='%s' was not found or was not accessible", id.String()),
		})
	} else {
		c.JSON(http.StatusOK, api.EnvironmentObject{
			Id:      (*environments)[0].Id,
			Display: (*environments)[0].Display,
		})
	}
}

func (r RouteHandlers) GetEnvironmentsV1(c *gin.Context, projectId api.ID) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.GetEnvironmentsWithResponse(context.Background(), &postgrest.GetEnvironmentsParams{ProjectId: equals(projectId)}); err != nil {
		state.Get().GetLogger().Debugf("[%s] request failed: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: "a pre-flight error occurred while processing the upstream request",
		})
	} else if response.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: fmt.Sprintf("error %d", response.StatusCode()),
		})
	} else if environments, err := parse[[]postgrest.Environments](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else {
		c.JSON(http.StatusOK, utils.ForEach(*environments, func(environment postgrest.Environments) api.EnvironmentObject {
			return api.EnvironmentObject{
				Id:      environment.Id,
				Display: environment.Display,
			}
		}))
	}
}

func (r RouteHandlers) CreateEnvironmentV1(c *gin.Context, projectId api.ID) {
	var req api.CreateEnvironmentV1JSONRequestBody
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &api.Error{
			Error:       "invalid request body",
			Description: err.Error(),
		})
	} else if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.PostEnvironmentsWithResponse(context.Background(), &postgrest.PostEnvironmentsParams{Prefer: preferFull[postgrest.PostEnvironmentsParamsPrefer]()}, postgrest.PostEnvironmentsApplicationVndPgrstObjectPlusJSONRequestBody{Display: req.Name, Id: uuid.New(), ProjectId: projectId}); err != nil {
		state.Get().GetLogger().Debugf("[%s] request failed: %v", c.Request.URL.Path, err)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "request failed",
			Description: "a pre-flight error occurred while processing the upstream request",
		})
	} else if response.StatusCode() == http.StatusConflict {
		c.JSON(http.StatusBadRequest, &api.Error{
			Error:       "duplicate",
			Description: "an object with this display-name already exists",
		})
	} else if response.StatusCode() != http.StatusCreated {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "bad response",
			Description: "an error occurred while processing the upstream response",
		})
	} else if environment, err := parseOne[postgrest.Environments](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else {
		c.JSON(http.StatusCreated, api.IDResponse{Id: environment.Id})
	}
}

func (r RouteHandlers) GetEnvironmentSecretsV1(c *gin.Context, environmentId api.ID) {
	//TODO implement me
	panic("implement me")
}
