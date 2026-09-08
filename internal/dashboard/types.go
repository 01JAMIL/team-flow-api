package dashboard

type kpiResponse struct {
	Workspaces     int64 `json:"workspaces"`
	ActiveProjects int64 `json:"activeProjects"`
	OpenTasks      int64 `json:"openTasks"`
	TeamMembers    int64 `json:"teamMembers"`
}
