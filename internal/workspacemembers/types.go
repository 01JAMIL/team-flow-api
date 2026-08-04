package workspacemembers

type addWorkspaceMemberPayload struct {
	UserID   string `json:"userId" binding:"required"`
	UserRole string `json:"userRole" binding:"required,oneof=ADMIN MEMBER"`
}
