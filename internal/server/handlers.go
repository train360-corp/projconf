package server

import (
	"github.com/gin-gonic/gin"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/train360-corp/projconf/internal/supabase"
	"net/http"
)

// RouteHandlers implements api.ServerInterface (generated).
type RouteHandlers struct{}

func (s *RouteHandlers) GetV1ClientsSelf(c *gin.Context) {

	sb := supabase.GetFromContext(c)
	if client, clientErr := sb.GetSelf(); clientErr != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":   "authentication failed",
			"details": clientErr.Error(),
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "secrets retrieval failed",
			"details": err.Error(),
		})
		return
	}

	if len(secrets) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "secrets retrieval failed",
			"details": "no secrets found",
		})
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
