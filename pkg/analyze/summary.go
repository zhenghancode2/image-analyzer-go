package analyze

type Summary struct {
	Architecture   string          `json:"architecture"`
	OS             string          `json:"os"`
	Env            []string        `json:"env"`
	OSInfo         string          `json:"os_info"`
	PythonPackages []string        `json:"python_packages"`
	Tools          map[string]bool `json:"tools"`
}

type AnalyzeOptions struct {
	CheckOSInfo         bool     `json:"check_os_info"`
	CheckPythonPackages bool     `json:"check_python_packages"`
	CheckCommonTools    bool     `json:"check_common_tools"`
	SpecificCommands    []string `json:"specific_commands"`
}
