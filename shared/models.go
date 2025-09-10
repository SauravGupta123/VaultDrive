package shared

type FileEvent struct {
	Name      string `json:"name"`
	Operation string `json:"operation"` // "create", "update", "delete"
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
