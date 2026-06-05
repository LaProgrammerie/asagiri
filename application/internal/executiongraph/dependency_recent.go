package executiongraph

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"gopkg.in/yaml.v3"
)

const recentFailureLimit = 50

// DefaultRecentFailuresLoader reads failures from the runtime store and graph event logs.
type DefaultRecentFailuresLoader struct{}

func (DefaultRecentFailuresLoader) RecentFlowFailures(ctx context.Context, repoRoot, product, flow string, limit int) ([]RecentFlowFailure, error) {
	_ = ctx
	flow = strings.TrimSpace(flow)
	if flow == "" || strings.TrimSpace(repoRoot) == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = recentFailureLimit
	}

	var out []RecentFlowFailure
	if store, err := runtime.Open(repoRoot); err == nil {
		defer func() { _ = store.Close() }()
		events, err := store.ListEvents(limit * 4)
		if err != nil {
			return nil, err
		}
		for _, ev := range events {
			if ev.Type != runtime.EventGraphNodeFailed {
				continue
			}
			if !flowMatches(ev.FlowID, flow) {
				continue
			}
			out = append(out, RecentFlowFailure{
				FlowID:    flow,
				GraphID:   stringPayload(ev.Payload, "graph_id"),
				NodeID:    stringPayload(ev.Payload, "node_id"),
				EventType: ev.Type,
				CreatedAt: ev.CreatedAt,
			})
			if len(out) >= limit {
				return out, nil
			}
		}
	}

	fromGraphs, err := recentFailuresFromGraphDirs(repoRoot, product, flow, limit)
	if err != nil {
		return out, err
	}
	seen := make(map[string]struct{}, len(out))
	for _, f := range out {
		seen[failureKey(f)] = struct{}{}
	}
	for _, f := range fromGraphs {
		key := failureKey(f)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, f)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func recentFailuresFromGraphDirs(repoRoot, product, flow string, limit int) ([]RecentFlowFailure, error) {
	graphsRoot := filepath.Join(repoRoot, ".asagiri", "graphs")
	entries, err := os.ReadDir(graphsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out []RecentFlowFailure
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		graphDir := filepath.Join(graphsRoot, entry.Name())
		metaFlow, metaProduct, err := graphMetadata(graphDir)
		if err != nil || !flowMatches(metaFlow, flow) {
			continue
		}
		if product != "" && metaProduct != "" && !strings.EqualFold(metaProduct, product) {
			continue
		}
		failures, err := failuresInGraphEvents(filepath.Join(graphDir, "events.jsonl"), flow, entry.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, failures...)
		if len(out) >= limit {
			return out[:limit], nil
		}
	}
	return out, nil
}

func graphMetadata(graphDir string) (flowID, productID string, err error) {
	for _, name := range []string{"execution-graph.yaml", "execution-graph.json"} {
		path := filepath.Join(graphDir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var doc struct {
			Flow    string `yaml:"flow" json:"flow"`
			Product string `yaml:"product" json:"product"`
		}
		if strings.HasSuffix(name, ".json") {
			if err := json.Unmarshal(raw, &doc); err != nil {
				return "", "", err
			}
		} else if err := yaml.Unmarshal(raw, &doc); err != nil {
			return "", "", err
		}
		return strings.TrimSpace(doc.Flow), strings.TrimSpace(doc.Product), nil
	}
	return "", "", fmt.Errorf("graph metadata not found in %s", graphDir)
}

func failuresInGraphEvents(path, flowID, graphID string) ([]RecentFlowFailure, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var out []RecentFlowFailure
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		eventType := firstString(row, "type", "event")
		if eventType != runtime.EventGraphNodeFailed {
			continue
		}
		out = append(out, RecentFlowFailure{
			FlowID:    flowID,
			GraphID:   graphID,
			NodeID:    firstString(row, "node_id", "node"),
			EventType: eventType,
		})
	}
	return out, sc.Err()
}

func flowMatches(stored, requested string) bool {
	stored = strings.TrimSpace(stored)
	requested = strings.TrimSpace(requested)
	if stored == "" || requested == "" {
		return false
	}
	return strings.EqualFold(stored, requested)
}

func failureKey(f RecentFlowFailure) string {
	return f.GraphID + "\x00" + f.NodeID + "\x00" + f.EventType
}

func stringPayload(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	v, ok := payload[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func firstString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := row[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
