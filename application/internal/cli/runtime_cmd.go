package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	runtimeapi "github.com/LaProgrammerie/asagiri/application/internal/runtime/api"
	"github.com/LaProgrammerie/asagiri/application/internal/skills"
	"github.com/spf13/cobra"
)

func newDaemonCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Runtime persistant local Asagiri",
	}
	var detach bool
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Démarrer le runtime local",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if *dryRun {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "dry-run: daemon start skipped")
				return nil
			}
			if detach {
				pid, err := runtime.StartDaemonDetached(root)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "daemon detached (pid %d)\n", pid)
				return nil
			}
			st, err := runtime.StartDaemon(root)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), runtime.FormatStatusPlain(st))
			return nil
		},
	}
	startCmd.Flags().BoolVar(&detach, "detach", false, "Lancer le worker en arrière-plan")

	var apiPort int
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Boucle worker runtime (long-running)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if *dryRun {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "dry-run: daemon run skipped")
				return nil
			}
			_, _ = runtime.StartDaemon(root)
			if apiPort > 0 {
				go func() {
					_ = runtimeapi.Serve(cmd.Context(), runtimeapi.Options{RepoRoot: root, Port: apiPort})
				}()
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "runtime API: http://%s\n", apiAddr(apiPort))
			}
			return runtime.RunWorker(cmd.Context(), root)
		},
	}
	runCmd.Flags().IntVar(&apiPort, "api-port", 0, "Démarrer l'API HTTP locale en parallèle (0 = désactivé)")

	var richStatus bool
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "État du runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			if richStatus {
				view, err := store.BuildStatusView()
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(cmd.OutOrStdout(), runtime.FormatStatusRich(view))
				return nil
			}
			st, err := store.Status()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), runtime.FormatStatusPlain(st))
			return nil
		},
	}
	statusCmd.Flags().BoolVar(&richStatus, "rich", false, "Affichage terminal enrichi (spec §24.20)")
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Arrêter le runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if *dryRun {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "dry-run: daemon stop skipped")
				return nil
			}
			if err := runtime.StopDaemon(root); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "runtime stopped")
			return nil
		},
	}
	cmd.AddCommand(startCmd, runCmd, statusCmd, stopCmd)
	return cmd
}

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Compétences réutilisables (.asagiri/skills)",
	}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les skills chargées",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			all, err := skills.LoadAll(root)
			if err != nil {
				return err
			}
			for _, s := range all {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%v\n", s.ID, s.Name, s.Scope)
			}
			return nil
		},
	}
	cmd.AddCommand(listCmd)
	return cmd
}

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memory",
		Short:   "Mémoire persistante runtime",
		Example: "  asa memory doctor\n  asa memory list --query \"checkout flow\"\n  asa memory reindex",
	}
	var scope, query string
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Lister les entrées mémoire",
		Example: "  asa memory list --scope session\n  asa memory list --query \"payment webhook\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := configureMemoryEmbedder(root); err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			eng := memory.NewEngine(store)
			var entries []runtime.MemoryEntry
			if query != "" {
				entries, err = eng.RetrieveByQuery(cmd.Context(), query, 50)
			} else {
				entries, err = eng.Retrieve(runtime.MemoryScope(scope), nil, 50)
			}
			if err != nil {
				return err
			}
			for _, e := range entries {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%.2f\t%s\t%s\n", e.Relevance, e.Scope, e.Summary)
			}
			return nil
		},
	}
	consolidateCmd := &cobra.Command{
		Use:     "consolidate",
		Short:   "Consolider les entrées proches",
		Example: "  asa memory consolidate",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			n, err := memory.NewEngine(store).Consolidate()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "consolidated: %d\n", n)
			return nil
		},
	}
	listCmd.Flags().StringVar(&scope, "scope", "", "Filtrer par scope")
	listCmd.Flags().StringVar(&query, "query", "", "Recherche sémantique (embeddings)")

	cmd.AddCommand(listCmd, consolidateCmd, newMemoryReindexCmd(), newMemoryDoctorCmd())
	return cmd
}

func newSessionCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Sessions d'ingénierie runtime",
	}
	var productID, flowID string
	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Créer une session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			sess, err := store.CreateSession(args[0], productID, flowID)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "session créée: %s (%s)\n", sess.ID, sess.Name)
			return nil
		},
	}
	createCmd.Flags().StringVar(&productID, "product", "", "Produit lié")
	createCmd.Flags().StringVar(&flowID, "flow", "", "Flow lié")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lister les sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			sessions, err := store.ListSessions()
			if err != nil {
				return err
			}
			for _, s := range sessions {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", s.ID, s.Name, s.Status)
			}
			return nil
		},
	}

	attachCmd := &cobra.Command{
		Use:   "attach <session>",
		Short: "Attacher le terminal à une session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			sess, err := store.GetSession(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "attached session: %s (%s) product=%s flow=%s\n",
				sess.Name, sess.ID, sess.ProductID, sess.FlowID)
			return nil
		},
	}

	var branchType string
	branchCmd := &cobra.Command{
		Use:   "branch <session>",
		Short: "Créer une branche exploratoire",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			b, err := store.CreateBranch(args[0], name, runtime.BranchType(branchType), "")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "branch créée: %s (%s)\n", b.ID, b.Name)
			return nil
		},
	}
	branchCmd.Flags().String("name", "", "Nom de la branche")
	branchCmd.Flags().StringVar(&branchType, "type", "flow", "Type de branche")

	graphCmd := &cobra.Command{
		Use:   "graph",
		Short: "Afficher le graphe sessions/branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			g, err := store.BuildStateGraph()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Sessions: %d\nBranches: %d\nRecent events: %d\n",
				len(g.Sessions), len(g.Branches), len(g.Events))
			for _, s := range g.Sessions {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  session %s (%s)\n", s.Name, s.Status)
			}
			for _, e := range g.Events {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  event %s %s\n", e.Type, e.CreatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}

	cmd.AddCommand(createCmd, listCmd, attachCmd, branchCmd, graphCmd)
	return cmd
}

func newRuntimeEventsCmd(dryRun *bool) *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Événements runtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			store, err := runtime.Open(root)
			if err != nil {
				return err
			}
			defer func() { _ = store.Close() }()
			if follow {
				since := time.Now().UTC()
				for i := 0; i < 3; i++ {
					events, err := store.ListEventsSince(since, 20)
					if err != nil {
						return err
					}
					for _, e := range events {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", e.CreatedAt.Format(time.RFC3339), e.Type, e.SessionID)
						since = e.CreatedAt
					}
					time.Sleep(200 * time.Millisecond)
				}
				return nil
			}
			events, err := store.ListEvents(30)
			if err != nil {
				return err
			}
			for _, e := range events {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", e.CreatedAt.Format(time.RFC3339), e.Type, e.SessionID)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&follow, "follow", false, "Suivre les nouveaux événements (poll court V1)")
	return cmd
}

func newRuntimeCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Couche runtime persistante",
	}
	var port int
	var socketPath string
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serveur HTTP local JSON (127.0.0.1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			if *dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dry-run: runtime serve on 127.0.0.1:%d\n", port)
				return nil
			}
			opts := runtimeapi.Options{RepoRoot: root, Port: port, SocketPath: socketPath}
			if socketPath != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "runtime API listening on unix://%s\n", socketPath)
				if port > 0 {
					go func() {
						_ = runtimeapi.Serve(cmd.Context(), opts)
					}()
				}
				return runtimeapi.ServeUnix(cmd.Context(), opts)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "runtime API listening on http://%s\n", apiAddr(port))
			return runtimeapi.Serve(cmd.Context(), opts)
		},
	}
	serveCmd.Flags().IntVar(&port, "port", 8765, "Port TCP (localhost only)")
	serveCmd.Flags().StringVar(&socketPath, "socket", "", "Socket Unix (ex: .asagiri/runtime/runtime.sock)")
	cmd.AddCommand(serveCmd, newRuntimeEventsCmd(dryRun))
	return cmd
}

func apiAddr(port int) string {
	if port <= 0 {
		port = 8765
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}
