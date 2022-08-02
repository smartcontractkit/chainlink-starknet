package integration_tests

import (
	"bytes"
	"text/template"
)

// P2PData holds the remote ip and the peer id and port
type P2PData struct {
	RemoteIP   string
	RemotePort string
	PeerID     string
}

// OCR2TaskJobSpec represents an OCR2 job that is given to other nodes, meant to communicate with the bootstrap node,
// and provide their answers
type OCR2TaskJobSpec struct {
	Name                  string            `toml:"name"`
	JobType               string            `toml:"type"`
	ContractID            string            `toml:"contractID"` // Address of the OCR contract/account(s)
	Relay                 string            `toml:"relay"`      // Name of blockchain relay to use
	P2pPeerID             string            `toml:"p2pPeerID"`
	PluginType            string            `toml:"pluginType"`            // Type of report plugin to use
	RelayConfig           map[string]string `toml:"relayConfig"`           // Relay spec object in stringified form
	P2pBootstrapPeers     []P2PData         `toml:"p2pBootstrapPeers"`     // P2P ID of the bootstrap node
	OCRKeyBundleID        string            `toml:"ocrKeyBundleID"`        // ID of this node's OCR key bundle
	TransmitterID         string            `toml:"transmitterID"`         // ID of address this node will use to transmit
	ObservationSource     string            `toml:"observationSource"`     // List of commands for the chainlink node
	JuelsPerFeeCoinSource string            `toml:"juelsPerFeeCoinSource"` // List of commands to fetch JuelsPerFeeCoin value (used to calculate ocr payments)
}

// Type returns the type of the job
func (o *OCR2TaskJobSpec) Type() string { return o.JobType }

// String representation of the job
func (o *OCR2TaskJobSpec) String() (string, error) {
	ocr2TemplateString := `type = "{{ .JobType }}"
schemaVersion                          = 1
name 																	 = "{{.Name}}"
relay																	 = "{{.Relay}}"
contractID		                         = "{{.ContractID}}"
{{if .P2pBootstrapPeers}}
p2pBootstrapPeers                      = [
  {{range $peer := .P2pBootstrapPeers}}
  "{{$peer.PeerID}}@{{$peer.RemoteIP}}:{{if $peer.RemotePort}}{{$peer.RemotePort}}{{else}}6690{{end}}",
  {{end}}
]
{{else}}
p2pBootstrapPeers                      = []
{{end}}
p2pPeerID								 = "{{.P2pPeerID}}"
{{if eq .JobType "offchainreporting2" }}
pluginType                             = "{{ .PluginType }}"
ocrKeyBundleID                         = "{{.OCRKeyBundleID}}"
transmitterID                     		 = "{{.TransmitterID}}"
observationSource                      = """
{{.ObservationSource}}
"""
[pluginConfig]
juelsPerFeeCoinSource                  = """
{{.JuelsPerFeeCoinSource}}
"""
{{end}}
[relayConfig]
{{range $key, $value := .RelayConfig}}
{{$key}} = "{{$value}}"
{{end}}`

	return marshallTemplate(o, "OCR2 Job", ocr2TemplateString)
}

// marshallTemplate Helper to marshall templates
func marshallTemplate(jobSpec interface{}, name, templateString string) (string, error) {
	var buf bytes.Buffer
	tmpl, err := template.New(name).Parse(templateString)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buf, jobSpec)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}
