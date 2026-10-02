// Package engine executes the declarative subset of FlowPilot workflow graphs.
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flowpilot/flowpilot/internal/domain"
)

type Event func(nodeID, eventType string, payload map[string]any)

// Run executes a validated DAG deterministically. It deliberately supports no
// user supplied code: transforms are declarative field projections only.
func Run(ctx context.Context, graph domain.Graph, input map[string]any, emit Event) (map[string]any, error) {
	nodes := map[string]domain.Node{}
	next := map[string][]domain.Edge{}
	for _, n := range graph.Nodes { nodes[n.ID] = n }
	for _, e := range graph.Edges { next[e.Source] = append(next[e.Source], e) }
	data := map[string]any{"trigger": input, "nodes": map[string]any{}}
	var queue []string
	for _, n := range graph.Nodes { if strings.HasSuffix(n.Type, "_trigger") { queue = append(queue, n.ID) } }
	result := map[string]any{}
	for len(queue) > 0 {
		id := queue[0]; queue = queue[1:]; n := nodes[id]
		emit(id, "node.started", nil)
		out, route, err := execute(ctx, n, data)
		if err != nil { emit(id, "node.failed", map[string]any{"error": err.Error()}); return nil, fmt.Errorf("node %s: %w", id, err) }
		data["nodes"].(map[string]any)[id] = out
		emit(id, "node.succeeded", map[string]any{"output": out})
		if n.Type == "workflow_result" { result = out }
		for _, e := range next[id] { if route == "" || e.SourceHandle == "" || e.SourceHandle == route { queue = append(queue, e.Target) } }
	}
	return result, nil
}

func execute(ctx context.Context, n domain.Node, data map[string]any) (map[string]any, string, error) {
	c := n.Config
	switch n.Type {
	case "manual_trigger", "webhook_trigger", "schedule_trigger":
		return data["trigger"].(map[string]any), "", nil
	case "set_variable", "json_transform":
		out := map[string]any{}
		for key, raw := range c { if key != "remove" { out[key] = resolve(raw, data) } }
		return out, "", nil
	case "condition":
		left := resolve(c["left"], data); right := resolve(c["right"], data); op, _ := c["operator"].(string)
		ok := compare(left, right, op)
		route := "false"; if ok { route = "true" }
		return map[string]any{"matched": ok}, route, nil
	case "delay":
		seconds, _ := c["seconds"].(float64); if seconds < 0 || seconds > 30 { return nil, "", errors.New("delay must be between 0 and 30 seconds") }
		select { case <-time.After(time.Duration(seconds*float64(time.Second))): return map[string]any{"delayedSeconds":seconds},"",nil; case <-ctx.Done(): return nil,"",ctx.Err() }
	case "http_request":
		raw, _ := resolve(c["url"], data).(string); if err := safeURL(raw); err != nil { return nil, "", err }
		method, _ := c["method"].(string); if method == "" { method="GET" }
		req, err := http.NewRequestWithContext(ctx, method, raw, nil); if err != nil { return nil,"",err }
		client:=http.Client{Timeout:10*time.Second, CheckRedirect: func(r *http.Request, _ []*http.Request) error { return safeURL(r.URL.String()) }}
		started:=time.Now(); resp,err:=client.Do(req);if err!=nil{return nil,"",err};defer resp.Body.Close()
		return map[string]any{"status":resp.StatusCode,"durationMs":time.Since(started).Milliseconds()},"",nil
	case "workflow_result":
		out:=map[string]any{};for k,v:=range c{out[k]=resolve(v,data)};return out,"",nil
	default: return nil,"",fmt.Errorf("unsupported node type %q",n.Type)
	}
}
func resolve(v any, data map[string]any) any { s,ok:=v.(string);if !ok||!strings.HasPrefix(s,"{{")||!strings.HasSuffix(s,"}}") {return v};parts:=strings.Split(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s,"{{"),"}}")),".");var x any=data;for _,p:=range parts{m,ok:=x.(map[string]any);if !ok{return nil};x=m[p]};return x }
func compare(a,b any,op string) bool { switch op {case "equals":return fmt.Sprint(a)==fmt.Sprint(b);case "not_equals":return fmt.Sprint(a)!=fmt.Sprint(b);case "exists":return a!=nil;case "contains":return strings.Contains(fmt.Sprint(a),fmt.Sprint(b));case "greater_than":return number(a)>number(b);case "less_than":return number(a)<number(b)};return false }
func number(v any)float64{var f float64;_ = json.Unmarshal([]byte(fmt.Sprint(v)),&f);return f}
func safeURL(raw string) error { u,e:=url.Parse(raw);if e!=nil||u.Scheme!="http"&&u.Scheme!="https"||u.Hostname()==""{return errors.New("invalid HTTP URL")};ips,e:=net.LookupIP(u.Hostname());if e!=nil{return e};for _,ip:=range ips{if ip.IsLoopback()||ip.IsPrivate()||ip.IsLinkLocalUnicast()||ip.IsUnspecified(){return errors.New("private or local HTTP targets are blocked")}};return nil }
