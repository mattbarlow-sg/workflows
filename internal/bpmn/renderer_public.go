package bpmn

// Public wrapper methods for renderer

// RenderDOT renders the process as GraphViz DOT format
func (r *Renderer) RenderDOT() string {
	return r.renderDOT()
}

// RenderMermaid renders the process as Mermaid diagram
func (r *Renderer) RenderMermaid() string {
	return r.renderMermaid()
}

// RenderText renders the process as text
func (r *Renderer) RenderText() string {
	return r.renderText()
}