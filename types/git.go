package types

// Platform represents a platform configuration
type Platform struct {
    Name        string      `json:"name"`
    ID          string      `json:"id"`
    PublicToken string      `json:"public_token"`
    URL         URL         `json:"URL"`
    URLStruc    URLStruc    `json:"URLStruc"`
    RawURLStruc RawURLStruc `json:"rawURLStruc"`
}

// URL represents the base URLs for a platform
type URL struct {
    Site []string `json:"site"`
    Raw  []string `json:"raw"`
}

// URLStruc represents the URL structure templates for a platform
type URLStruc struct {
    Site         string `json:"site"`
    CommitFolder string `json:"commit_folder"`
    CommitFile   string `json:"commit_file"`
    BranchFolder string `json:"branch_folder"`
    BranchFile   string `json:"branch_file"`
}

// RawURLStruc represents the raw URL structure templates for a platform
type RawURLStruc struct {
    Site   string `json:"site"`
    Commit string `json:"commit"`
    Branch string `json:"branch"`
}