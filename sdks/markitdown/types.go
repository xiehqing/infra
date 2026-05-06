package markitdown

type ClientOption struct {
	URL       string `json:"url" yaml:"url" mapstructure:"url"`
	Timeout   int    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	PrintLog  bool   `json:"printLog" yaml:"print-log" mapstructure:"print-log"`
	Sensitive bool   `json:"sensitive" yaml:"sensitive" mapstructure:"sensitive"`
}
