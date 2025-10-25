package config

type GitHubConf struct {
	GitHubToken string `json:"githubToken"`
	Owner       string `json:"owner"` // GitHub repository owner
	Repo        string `json:"repo"`  // GitHub repository name
}

type GitlabConf struct {
	GitLabURL    string `json:"url"`
	PrivateToken string `json:"privateToken"`
	ProjectID    string `json:"projectID"`
}

type JenkinsConf struct {
	Username   string `json:"username"`
	ApiToken   string `json:"apitoken"`
	JenkinsURL string `json:"url"`
	ProjectID  string `json:"projectID"`
}

type Config struct {
	GitHubConf  GitHubConf  `json:"githubConf"`
	GitlabConf  GitlabConf  `json:"gitlabConf"`
	JenkinsConf JenkinsConf `json:"jenkinsConf"`
	BinDir      string      `json:"binDir"`
}
