package shared

import "time"

type FileEvent struct {
	Name      string `json:"name"`
	Operation string `json:"operation"` // "create", "update", "delete"
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type FileMetadata struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
