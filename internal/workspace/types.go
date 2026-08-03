package workspace

type createWorkspacePayload struct {
	WorkspaceName string `json:"workspaceName" binding:"required"`
	Description   string `json:"description" binding:"required"`
	UserID        string `json:"userID,omitempty"`
}
