package managerresp

type CategoryPermissionResponse struct {
	Value string                       `json:"value"`
	Label string                       `json:"label"`
	Child []CategoryPermissionResponse `json:"child"`
}
