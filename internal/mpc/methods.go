package mpc

// GetNodeByID returns the node with the given ID, or nil if not found
func (m *MPC) GetNodeByID(id string) *Node {
	for i := range m.Nodes {
		if m.Nodes[i].ID == id {
			return &m.Nodes[i]
		}
	}
	return nil
}

// GetCompletedSubtaskCount returns the number of completed subtasks
func (n *Node) GetCompletedSubtaskCount() int {
	count := 0
	for _, subtask := range n.Subtasks {
		if subtask.Completed {
			count++
		}
	}
	return count
}

// GetCompletionPercentage returns the percentage of completed subtasks
func (n *Node) GetCompletionPercentage() float64 {
	if len(n.Subtasks) == 0 {
		return 0
	}
	completed := n.GetCompletedSubtaskCount()
	return float64(completed) / float64(len(n.Subtasks)) * 100
}
