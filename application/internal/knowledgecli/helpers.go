package knowledgecli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func osGetwdMust() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func printQueryResult(out interface{ Write([]byte) (int, error) }, result knowledge.GraphQueryResult) {
	_, _ = fmt.Fprintf(out, "Nodes (%d)\n", len(result.Nodes))
	for _, n := range result.Nodes {
		_, _ = fmt.Fprintf(out, "  %s  %s\n", n.ID, n.Name)
	}
	_, _ = fmt.Fprintf(out, "Edges (%d)\n", len(result.Edges))
	for _, e := range result.Edges {
		_, _ = fmt.Fprintf(out, "  %s  %s -> %s\n", e.Type, e.From, e.To)
	}
}
