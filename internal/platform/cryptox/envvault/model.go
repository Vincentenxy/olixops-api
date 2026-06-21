package envvault

import "time"

const (
	EnvDev  = "dev"
	EnvTest = "test"
	ENVStag = "stag"
	EnvProd = "prod"
)

type SecretsListRequest struct {
	ProjectId  string
	FolderCode string
	Key        string
	EnvList    []string
}

type SecretsListResponse struct {
	Comment     string                 `json:"comment"`
	Key         string                 `json:"key"`
	ProjectCode string                 `json:"projectCode"`
	EnvItems    map[string]*SecretItem `json:",inline"`
}

type SecretItem struct {
	FolderId  string    `json:"folderId"`
	Value     string    `json:"value"`
	Version   string    `json:"version"`
	Comment   string    `json:"comment"`
	UpdatedAt time.Time `json:"updatedAt"`
}
