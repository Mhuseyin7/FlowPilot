package domain

import "fmt"

var known = map[string]bool{"manual_trigger":true,"webhook_trigger":true,"schedule_trigger":true,"http_request":true,"json_transform":true,"set_variable":true,"condition":true,"delay":true,"workflow_result":true}
func ValidateGraph(g Graph) []string {
 var errs []string; ids:=map[string]bool{}; triggers:=0
 for _, n:=range g.Nodes { if n.ID=="" || ids[n.ID] { errs=append(errs,"node IDs must be unique and non-empty") }; ids[n.ID]=true; if !known[n.Type] { errs=append(errs,fmt.Sprintf("node %s has unsupported type %s",n.ID,n.Type)) }; if n.Type=="webhook_trigger"&&fmt.Sprint(n.Config["webhookId"])=="" {errs=append(errs,fmt.Sprintf("webhook trigger %s requires webhookId",n.ID))}; if n.Type=="http_request"&&fmt.Sprint(n.Config["url"])=="" {errs=append(errs,fmt.Sprintf("HTTP node %s requires url",n.ID))}; if n.Type=="manual_trigger"||n.Type=="webhook_trigger"||n.Type=="schedule_trigger" { triggers++ } }
 if triggers!=1 { errs=append(errs,"a workflow must contain exactly one trigger") }
 indegree:=map[string]int{}; adj:=map[string][]string{}
 for _,e:=range g.Edges { if !ids[e.Source]||!ids[e.Target] { errs=append(errs,"edge references an unknown node"); continue }; if e.Source==e.Target { errs=append(errs,"self edges are not allowed") }; indegree[e.Target]++; adj[e.Source]=append(adj[e.Source],e.Target) }
 queue:=[]string{}; for id:=range ids {if indegree[id]==0 {queue=append(queue,id)}}; seen:=0; for len(queue)>0 { n:=queue[0];queue=queue[1:];seen++;for _,to:=range adj[n] {indegree[to]--;if indegree[to]==0 {queue=append(queue,to)}} }; if seen!=len(ids) {errs=append(errs,"cycles are not supported")}; return errs
}
