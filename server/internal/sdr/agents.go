package sdr

// SDR pipeline agent names — single source of truth.
// The seed handler (sdr_seed.go), the engine (engine.go), and migrations
// MUST all use these constants. Changing a name here requires updating:
//   1. sdr_seed.go sdrAgentDefs[].Name
//   2. engine.go onTaskCompleted switch
//   3. Migration files that reference agent names (053, 059)
//   4. Production DB via a new migration (never edit applied migrations)
const (
	AgentHunter       = "Hunter"
	AgentQualificador = "Qualificador"
	AgentCopywriter   = "Copywriter"
	AgentCloser       = "Closer"
	AgentNurture      = "Nurture"
)

// PipelineOrder defines the SDR handoff chain.
var PipelineOrder = []string{
	AgentHunter,
	AgentQualificador,
	AgentCopywriter,
	AgentCloser,
	AgentNurture,
}

// NextAgent returns the next agent in the pipeline, or "" if none.
func NextAgent(name string) string {
	for i, a := range PipelineOrder {
		if a == name && i+1 < len(PipelineOrder) {
			return PipelineOrder[i+1]
		}
	}
	return ""
}