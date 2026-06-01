package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cristian-scherer/memoos/internal/app"
	"github.com/cristian-scherer/memoos/internal/models"
	"github.com/cristian-scherer/memoos/internal/util"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
		os.Exit(1)
	}
	defer application.Stop()

	if err := application.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start app: %v\n", err)
		os.Exit(1)
	}

	ctx := application.Context()

	switch args[0] {
	case "save":
		handleSave(ctx, application, args[1:])
	case "search":
		handleSearch(ctx, application, args[1:])
	case "list":
		handleList(ctx, application, args[1:])
	case "get":
		handleGet(ctx, application, args[1:])
	case "delete":
		handleDelete(ctx, application, args[1:])
	case "categories":
		handleCategories(ctx, application, args[1:])
	case "health":
		handleHealth(ctx, application)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("MemoOS CLI - Semantic memory for AI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  memoos-cli <command> [command-options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  save        Save a memory")
	fmt.Println("  search      Search memories")
	fmt.Println("  list        List memories")
	fmt.Println("  get         Get a memory by ID")
	fmt.Println("  delete      Delete a memory by ID")
	fmt.Println("  categories  List categories for a project")
	fmt.Println("  health      Check system health")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  memoos-cli save --cwd /home/user/project --category payments --content 'Refund Pix uses e2eid'")
	fmt.Println("  memoos-cli search --cwd /home/user/project --query 'how does refund work?'")
	fmt.Println("  memoos-cli list --cwd /home/user/project --category payments")
}

func handleSave(ctx context.Context, app *app.App, args []string) {
	flags := flag.NewFlagSet("save", flag.ExitOnError)
	cwd := flags.String("cwd", getCWD(), "Current working directory")
	category := flags.String("category", "", "Memory category (optional)")
	content := flags.String("content", "", "Memory content")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	if *content == "" {
		fmt.Println("Error: --content is required")
		os.Exit(1)
	}

	var categoryPtr *string
	if *category != "" {
		categoryPtr = category
	}

	input := models.MemoryInput{
		CWD:      *cwd,
		Category: categoryPtr,
		Content:  *content,
	}

	mem, err := app.MemService.Save(ctx, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	categoryDisplay := "<none>"
	if mem.Category != nil {
		categoryDisplay = *mem.Category
	}
	fmt.Printf("Memory saved (ID: %d, Project: %s, Category: %s)\n", mem.ID, mem.Project, categoryDisplay)
}

func handleSearch(ctx context.Context, app *app.App, args []string) {
	flags := flag.NewFlagSet("search", flag.ExitOnError)
	cwd := flags.String("cwd", getCWD(), "Current working directory")
	query := flags.String("query", "", "Search query")
	category := flags.String("category", "", "Filter by category")
	limit := flags.Int("limit", 10, "Maximum results")
	minScore := flags.Float64("min-score", 0.0, "Minimum similarity score (0-1)")
	maxDistance := flags.Float64("max-distance", 0.0, "Maximum euclidean distance")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	if *query == "" {
		fmt.Println("Error: --query is required")
		os.Exit(1)
	}

	var categoryPtr *string
	if *category != "" {
		str := *category
		categoryPtr = &str
	}

	input := models.SearchInput{
		CWD:         *cwd,
		Query:       *query,
		Category:    categoryPtr,
		Limit:       *limit,
		MinScore:    float32(*minScore),
		MaxDistance: float32(*maxDistance),
	}

	results, err := app.MemService.Search(ctx, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d results\n", len(results))
	for i, r := range results {
		fmt.Printf("\n%d. [%.2f] %s\n", i+1, r.Score, r.Memory.Content)
		categoryDisplay := "<none>"
		if r.Memory.Category != nil {
			categoryDisplay = *r.Memory.Category
		}
		fmt.Printf("   Project: %s | Category: %s | ID: %d\n", r.Memory.Project, categoryDisplay, r.Memory.ID)
	}
}

func handleList(ctx context.Context, app *app.App, args []string) {
	flags := flag.NewFlagSet("list", flag.ExitOnError)
	cwd := flags.String("cwd", getCWD(), "Current working directory")
	category := flags.String("category", "", "Filter by category")
	limit := flags.Int("limit", 20, "Maximum results")
	offset := flags.Int("offset", 0, "Pagination offset")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	projName := ""
	if *cwd != "" {
		projName = util.ResolvePath(*cwd)
	}

	var categoryPtr *string
	if *category != "" {
		str := *category
		categoryPtr = &str
	}

	filter := models.MemoryFilter{
		Project:  projName,
		Category: categoryPtr,
	}

	pagination := &models.Pagination{
		Page:     (*offset / *limit) + 1,
		PageSize: *limit,
	}

	memories, err := app.MemService.List(ctx, filter, pagination)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d memories\n", len(memories))
	for i, m := range memories {
		fmt.Printf("\n%d. %s\n", i+1, m.Content)
		categoryDisplay := "<none>"
		if m.Category != nil {
			categoryDisplay = *m.Category
		}
		fmt.Printf("   Project: %s | Category: %s | ID: %d | Created: %s\n", m.Project, categoryDisplay, m.ID, m.CreatedAt.Format("2006-01-02 15:04"))
	}
}

func handleGet(ctx context.Context, app *app.App, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: memory ID is required")
		os.Exit(1)
	}

	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid memory ID: %v\n", err)
		os.Exit(1)
	}

	mem, err := app.MemService.Get(ctx, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ID: %d\n", mem.ID)
	fmt.Printf("Project: %s\n", mem.Project)
	categoryDisplay := "<none>"
	if mem.Category != nil {
		categoryDisplay = *mem.Category
	}
	fmt.Printf("Category: %s\n", categoryDisplay)
	fmt.Printf("Created: %s\n", mem.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", mem.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nContent:\n%s\n", mem.Content)

	if len(mem.Metadata) > 0 {
		fmt.Printf("\nMetadata:\n")
		b, _ := json.MarshalIndent(mem.Metadata, "  ", "  ")
		fmt.Printf("  %s\n", string(b))
	}
}

func handleDelete(ctx context.Context, app *app.App, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: memory ID is required")
		os.Exit(1)
	}

	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid memory ID: %v\n", err)
		os.Exit(1)
	}

	if err := app.MemService.Delete(ctx, id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Memory %d deleted\n", id)
}

func handleCategories(ctx context.Context, app *app.App, args []string) {
	flags := flag.NewFlagSet("categories", flag.ExitOnError)
	cwd := flags.String("cwd", getCWD(), "Current working directory")

	if err := flags.Parse(args); err != nil {
		os.Exit(1)
	}

	projName := ""
	if *cwd != "" {
		projName = util.ResolvePath(*cwd)
	}

	categories, err := app.MemService.ListCategories(ctx, projName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if projName == "" {
		fmt.Println("Categories:")
	} else {
		fmt.Printf("Categories for project %s:\n", projName)
	}

	for _, cat := range categories {
		fmt.Printf("  - %s\n", cat)
	}
}

func handleHealth(ctx context.Context, app *app.App) {
	if err := app.Health(ctx); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Health: OK")
	fmt.Printf("Embedder: %s (dimension: %d)\n", app.Embedder.Name(), app.Embedder.Dimension())
}

func getCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}
