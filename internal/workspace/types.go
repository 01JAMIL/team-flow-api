package workspace

type createWorkspacePayload struct {
	WorkspaceName string `json:"workspaceName" binding:"required"`
	Description   string `json:"description" binding:"required"`
	UserID        string `json:"userID,omitempty"`
}

type updateWorkspacePayload struct {
	ID            string `json:"id,omitempty"`
	UserID        string `json:"userID,omitempty"`
	WorkspaceName string `json:"workspaceName,omitempty"`
	Description   string `json:"description,omitempty"`
}
