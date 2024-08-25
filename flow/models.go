package flow

type FlowData struct {
	Edges []Edge `json:"edges"`
	Nodes []Node `json:"nodes"`
	ID    string `json:"id"`
}

type Edge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Deletable bool   `json:"deletable"`
}

type Node struct {
	// Add fields as needed
	ID       string `json:"id"`
	Type     string `json:"type"`
	Data     any    `json:"data"`
	Position any    `json:"position"`
	Measured any    `json:"measured"`
}

type NodeResult struct {
	Type     string
	Continue bool
}
