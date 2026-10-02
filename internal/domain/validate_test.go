package domain

import "testing"

func TestValidateGraph(t *testing.T) {
	valid := Graph{Nodes: []Node{{ID:"trigger", Type:"manual_trigger"}, {ID:"result", Type:"workflow_result"}}, Edges: []Edge{{Source:"trigger", Target:"result"}}}
	if got := ValidateGraph(valid); len(got) != 0 { t.Fatalf("valid graph rejected: %v", got) }
	cycle := Graph{Nodes: []Node{{ID:"trigger", Type:"manual_trigger"}, {ID:"result", Type:"workflow_result"}}, Edges: []Edge{{Source:"trigger", Target:"result"},{Source:"result",Target:"trigger"}}}
	if got := ValidateGraph(cycle); len(got) == 0 { t.Fatal("cycle accepted") }
}
