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

func (r RouteHandlers) GetProjectsV1(c *gin.Context) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.GetProjectsWithResponse(context.Background(), &postgrest.GetProjectsParams{}); err != nil {
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
	} else if projects, err := parse[[]postgrest.Projects](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else {
		c.JSON(http.StatusOK, utils.ForEach(*projects, func(project postgrest.Projects) api.ProjectObject {
			return api.ProjectObject{
				Id:      project.Id,
				Display: project.Display,
			}
		}))
	}
}

func (r RouteHandlers) GetProjectV1(c *gin.Context, projectId api.ID) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.GetProjectsWithResponse(context.Background(), &postgrest.GetProjectsParams{Id: equals(projectId.String())}); err != nil {
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
	} else if projects, err := parse[[]postgrest.Projects](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else if len(*projects) == 0 {
		c.JSON(http.StatusNotFound, &api.Error{
			Error:       "not found",
			Description: fmt.Sprintf("a project with id='%s' was not found or was not accessible", projectId.String()),
		})
	} else {
		c.JSON(http.StatusOK, api.ProjectObject{
			Id:      (*projects)[0].Id,
			Display: (*projects)[0].Display,
		})
	}
}

func (r RouteHandlers) CreateProjectV1(c *gin.Context) {
	var req api.CreateProjectV1JSONRequestBody
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &api.Error{
			Error:       "invalid request body",
			Description: err.Error(),
		})
	} else if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.PostProjectsWithResponse(context.Background(), &postgrest.PostProjectsParams{Prefer: preferFull}, postgrest.PostProjectsApplicationVndPgrstObjectPlusJSONRequestBody{Display: req.Name, Id: uuid.New()}); err != nil {
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
	} else if project, err := parseOne[postgrest.Projects](response.Body); err != nil {
		state.Get().GetLogger().Debugf("[%d] %s", response.StatusCode(), response.Body)
		c.JSON(http.StatusInternalServerError, &api.Error{
			Error:       "unable to parse response",
			Description: "an error occurred while processing the upstream response",
		})
	} else {
		c.JSON(http.StatusCreated, api.IDResponse{Id: project.Id})
	}
}

func (r RouteHandlers) DeleteProjectV1(c *gin.Context, projectId api.ID) {
	if supabase, err := r.postgrest(c); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else if response, err := supabase.DeleteProjectsWithResponse(context.Background(), &postgrest.DeleteProjectsParams{Id: equals(projectId.String())}); err != nil {
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
