package provisionersdk

import "fmt"

var (
	// A mapping of operating-system ($GOOS) to architecture ($GOARCH)
	// to agent install and run script. ${DOWNLOAD_URL} is replaced
	// with strings.ReplaceAll() when being consumed.
	agentScripts = map[string]map[string]string{
		"windows": {
			"amd64": `
$ProgressPreference = "SilentlyContinue"
Invoke-WebRequest -Uri ${ACCESS_URL}bin/coder-windows-amd64.exe -OutFile $env:TEMP\coder.exe
$env:CODER_URL = "${ACCESS_URL}"
Start-Process -FilePath $env:TEMP\coder.exe -ArgumentList "workspaces","agent" -PassThru
`,
		},
		"linux": {
			"amd64": `
#!/usr/bin/env sh
set -eu pipefail
BINARY_LOCATION=$(mktemp -d)/coder
curl -fsSL ${ACCESS_URL}bin/coder-linux-amd64 -o $BINARY_LOCATION
chmod +x $BINARY_LOCATION
export CODER_URL="${ACCESS_URL}"
exec $BINARY_LOCATION workspaces agent
`,
		},
		"darwin": {
			"amd64": `
#!/usr/bin/env sh
set -eu pipefail
BINARY_LOCATION=$(mktemp -d)/coder
curl -fsSL ${ACCESS_URL}bin/coder-darwin-amd64 -o $BINARY_LOCATION
chmod +x $BINARY_LOCATION
export CODER_URL="${ACCESS_URL}"
exec $BINARY_LOCATION workspaces agent
`,
		},
	}
)

// AgentScriptEnv returns a key-pair of scripts that are consumed
// by the Coder Terraform Provider. See:
// https://github.com/coder/terraform-provider-coder/blob/main/internal/provider/provider.go#L97
func AgentScriptEnv() map[string]string {
	env := map[string]string{}
	for operatingSystem, scripts := range agentScripts {
		for architecture, script := range scripts {
			env[fmt.Sprintf("CODER_AGENT_SCRIPT_%s_%s", operatingSystem, architecture)] = script
		}
	}
	return env
}
