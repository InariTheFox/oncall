package models

type Version struct {
	Commit  string `json:"commit"`
	URL     string `json:"url"`
	Version string `json:"version"`
	// Arch contains architecture metadata.
	Arch map[string]ArchMeta `json:"arch"`
}

type ArchMeta struct {
	SHA256 string `json:"sha256"`
}
