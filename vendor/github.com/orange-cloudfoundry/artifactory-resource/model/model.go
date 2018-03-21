package model

type Source struct {
	Url       string `json:"url"`
	User      string `json:"user"`
	Password  string `json:"password"`
	ApiKey    string `json:"apiKey"`
	SshKey    string `json:"ssh_key"`
	Pattern   string `json:"pattern"`
	Props     string `json:"props"`
	Recursive bool   `json:"recursive"`
	Flat      bool   `json:"flat"`
	Regexp    bool   `json:"regexp"`
	LogLevel  string `json:"log_level"`
	CACert    string `json:"ca_cert"`
}
type InParams struct {
	Filename      string `json:"filename"`
	Version       string `json:"version"`
	Notflat       bool   `json:"not_flat"`
	Threads       int    `json:"threads"`
	MinSplit      int    `json:"min_split"`
	SplitCount    int    `json:"split_count"`
	PropsFilename string `json:"props_filename"`
}
type OutParams struct {
	Target         string `json:"target"`
	Source         string `json:"source"`
	Threads        int    `json:"threads"`
	ExplodeArchive bool   `json:"explode_archive"`
	Props          string `json:"props"`
	PropsFromFile  string `json:"props_from_file"`
}
