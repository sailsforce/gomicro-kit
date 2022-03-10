package models

type Heartbeat struct {
	RequestID      string `json:"requestId"`
	DatabaseOnline bool   `json:"databaseOnline"`
	AppName        string `json:"appName"`
	ReleaseDate    string `json:"releaseCreatedAt"`
	ReleaseVersion string `json:"releaseVersion"`
	Slug           string `json:"slugCommit"`
	Message        string `json:"message"`
}

type ServicePoolStatus struct {
	RequestID string      `json:"requestId"`
	Services  []Heartbeat `json:"services"`
}
