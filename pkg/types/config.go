package types

// Config represents persistent user configuration
type Config struct {
	GithubToken  string `json:"github_token,omitempty"`
	DownloadPath string `json:"download_path,omitempty"`
	ASCIIMode    bool   `json:"ascii_mode,omitempty"`
	Workers      int    `json:"workers,omitempty"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() Config {
	return Config{
		DownloadPath: ".",
		ASCIIMode:    false,
		Workers:      5,
	}
}

// Merge merges another config into this one, overwriting non-empty values
func (c *Config) Merge(other Config) {
	if other.GithubToken != "" {
		c.GithubToken = other.GithubToken
	}
	if other.DownloadPath != "" {
		c.DownloadPath = other.DownloadPath
	}
	if other.ASCIIMode {
		c.ASCIIMode = other.ASCIIMode
	}
	if other.Workers > 0 {
		c.Workers = other.Workers
	}
}
