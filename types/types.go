package types

// Application-wide constants
const (
	AppName = "JDUB Custom LLM Frontend"
)

// TemplateData holds data for rendering HTML templates
type TemplateData struct {
	Title         string
	Instructions  string
	Response      string
	Models        []string
	SelectedModel string
}

// SandboxEnv holds VM-related environment variables
type SandboxEnv struct {
	VMName     string
	VMUsername string
	VMPassword string
}

type SessionState struct {
	IntegrityScore float64
	ScoredRows     map[string]struct{}
	DatasetScores  map[string]float64
}
