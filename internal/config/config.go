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

type MongoConf struct {
	URL      string `json:"url"`
	Database string `json:"database"`
}

type Config struct {
	GitHubConf  GitHubConf  `json:"githubConf"`
	GitlabConf  GitlabConf  `json:"gitlabConf"`
	JenkinsConf JenkinsConf `json:"jenkinsConf"`
	MongoConf   MongoConf   `json:"mongoConf"`
}
