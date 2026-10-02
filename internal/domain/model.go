package domain

import "time"

type Graph struct { Nodes []Node `json:"nodes"`; Edges []Edge `json:"edges"` }
type Node struct { ID string `json:"id"`; Type string `json:"type"`; Config map[string]any `json:"config"` }
type Edge struct { Source string `json:"source"`; Target string `json:"target"`; SourceHandle string `json:"sourceHandle,omitempty"` }
type Workflow struct { ID string `json:"id"`; Name string `json:"name"`; Description string `json:"description"`; Status string `json:"status"`; Draft Graph `json:"draft"`; DraftRevision int64 `json:"draftRevision"`; PublishedVersion *int `json:"publishedVersion"`; CreatedAt time.Time `json:"createdAt"`; UpdatedAt time.Time `json:"updatedAt"` }
type Execution struct { ID string `json:"id"`; WorkflowID string `json:"workflowId"`; WorkflowVersion int `json:"workflowVersion"`; TriggerType string `json:"triggerType"`; Status string `json:"status"`; Input map[string]any `json:"input"`; Result any `json:"result,omitempty"`; Error *string `json:"error,omitempty"`; CreatedAt time.Time `json:"createdAt"`; StartedAt *time.Time `json:"startedAt,omitempty"`; FinishedAt *time.Time `json:"finishedAt,omitempty"` }
