package launch

import (
	"path"
	"strings"
)

type Process struct {
	Type        string   `toml:"type" json:"type"`
	Command     string   `toml:"command" json:"command"`
	Args        []string `toml:"args" json:"args"`
	Direct      bool     `toml:"direct" json:"direct"`
	BuildpackID string   `toml:"buildpack-id" json:"buildpackID"`
}

type Metadata struct {
	Processes  []Process   `toml:"processes" json:"processes"`
	Buildpacks []Buildpack `toml:"buildpacks" json:"buildpacks"`
}

type Buildpack struct {
	API string `toml:"api"`
	ID  string `toml:"id"`
}

type Env interface {
	AddRootDir(baseDir string) error
	AddEnvDir(envDir string) error
	List() []string
	Get(string) string
}

func EscapeID(id string) string {
	return strings.Replace(id, "/", "_", -1)
}

func GetMetadataFilePath(layersDir string) string {
	return path.Join(layersDir, "config", "metadata.toml")
}
