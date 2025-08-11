package server

import (
	"github.com/gin-gonic/gin"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

// RouteHandlers implements api.ServerInterface (generated).
type RouteHandlers struct{}

func (s *RouteHandlers) GetV1ProjectsProjectIdEnvironmentsEnvironmentIdSecrets(c *gin.Context, projectId openapitypes.UUID, environmentId openapitypes.UUID) {
	//TODO implement me
	panic("implement me")
}
