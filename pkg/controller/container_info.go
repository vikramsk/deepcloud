package controller

// RegistryType enumerates over the supported
// Container Registry Types.
type RegistryType string

const (
	// RegistryTypeDocker represents the Docker
	// container registry.
	RegistryTypeDocker = "Docker"
)

// ContainerInfo defines the attributes
// of the Container image. This could be
// stored at any container registry.
type ContainerInfo struct {
	ImageURL string `json:"imageUrl"`

	RegistryName RegistryType `json:"registryName"`

	ProjectInfo Project `json:"projectInfo"`
}

// Project defines the attributes of a
// User's DeepCloud project.
type Project struct {
	UserID string `json:"userID"`

	ProjectName string `json:"projectName"`
}
