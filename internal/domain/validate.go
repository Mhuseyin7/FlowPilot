package domain

import (
	"fmt"
	"strings"
)

var known = map[string]bool{
	"manual_trigger": true, "webhook_trigger": true, "schedule_trigger": true,
	"http_request": true, "json_transform": true, "set_variable": true,
	"condition": true, "delay": true, "workflow_result": true,
}

func ValidateGraph(g Graph) []string {
	var errs []string
	ids := map[string]bool{}
	triggers := 0
	for _, node := range g.Nodes {
		if node.ID == "" || ids[node.ID] {
			errs = append(errs, "node IDs must be unique and non-empty")
		}
		ids[node.ID] = true
		if !known[node.Type] {
			errs = append(errs, fmt.Sprintf("node %s has unsupported type %s", node.ID, node.Type))
		}
		if node.Type == "webhook_trigger" && !requiredString(node.Config, "webhookId") {
			errs = append(errs, fmt.Sprintf("webhook trigger %s requires webhookId", node.ID))
		}
		if node.Type == "http_request" && !requiredString(node.Config, "url") {
			errs = append(errs, fmt.Sprintf("HTTP node %s requires url", node.ID))
		}
		if node.Type == "manual_trigger" || node.Type == "webhook_trigger" || node.Type == "schedule_trigger" {
			triggers++
		}
	}
	if triggers != 1 {
		errs = append(errs, "a workflow must contain exactly one trigger")
	}

	indegree := map[string]int{}
	adj := map[string][]string{}
	edges := map[string]bool{}
	for _, edge := range g.Edges {
		if !ids[edge.Source] || !ids[edge.Target] {
			errs = append(errs, "edge references an unknown node")
			continue
		}
		key := edge.Source + "\x00" + edge.Target + "\x00" + edge.SourceHandle
		if edges[key] {
			errs = append(errs, "duplicate edges are not allowed")
			continue
		}
		edges[key] = true
		if edge.Source == edge.Target {
			errs = append(errs, "self edges are not allowed")
		}
		indegree[edge.Target]++
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
	}

	queue := []string{}
	for id := range ids {
		if indegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	seen := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		seen++
		for _, target := range adj[node] {
			indegree[target]--
			if indegree[target] == 0 {
				queue = append(queue, target)
			}
		}
	}
	if seen != len(ids) {
		errs = append(errs, "cycles are not supported")
	}
	return errs
}

func requiredString(config map[string]any, key string) bool {
	value, ok := config[key].(string)
	return ok && strings.TrimSpace(value) != ""
}
