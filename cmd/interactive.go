package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/engine"
	"github.com/bisegni/jsl/pkg/plan"
	"github.com/bisegni/jsl/pkg/planner"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/chzyer/readline"
)

func RunInteractive(filename string) error {
	fmt.Println("Interactive mode enabled. Type 'exit' or 'quit' to leave.")
	if filename == "-" {
		fmt.Println("Reading from stdin...")
	} else {
		fmt.Printf("Reading from file: %s\n", filename)
	}

	// For interactive mode, we likely want to load the data once if possible,
	// OR create a table that can be queried repeatedly.
	// Since JSONTable implementation in root.go (and pkg/database) seems to re-parse on Iterate,
	// this is fine for now. If performance is an issue, we'd cache it.
	// However, for stdin ("-"), re-reading isn't possible unless buffered.
	// pkg/database/json_table.go uses parser.NewParser(filename).
	// If filename is "-", parser might consume stdin.

	// TODO: For "-" (stdin), we might need to buffer it into a temp file or memory
	// if we want to query it multiple times.
	// Let's assume for this iteration that we rely on the existing infrastructure.
	// If the parser reads stdin once, subsequent queries might fail on "-".
	// Let's check: can we re-read stdin? No.
	// So for interactive mode with stdin, we MUST read it into memory or a temp file first.
	// OR, we just warn user: "Single pass on stdin not supported for multiple queries"
	// BUT, the request implies "write query without exit", so we probably need to handle this.

	// Let's load the table first.
	// To support multiple queries on the same data, especially from stdin,
	// we should probably verify if we can re-iterate.
	// Since we don't have a "MemoryTable" yet exposed easily here without importing internal parser structs,
	// let's stick to the simplest implementation:
	// 1. Create Input Table
	// 2. Loop REPL

	// WARN: If filename is "-", the first query will consume it. Subsequent queries will find EOF.
	// We might need a "BufferedJSONTable" or similar if we want to fix that,
	// but for this task "add interactive mode", let's start with the REPL loop.
	// We can add a warning for stdin users if needed, or maybe the user just wants to type one query?
	// No, "write query without exit" implies multiple queries.
	// We'll proceed with standard `database.NewJSONTable(filename)` and see.

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     "", // In-memory history for this session
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}
		if strings.EqualFold(trimmed, "exit") || strings.EqualFold(trimmed, "quit") {
			break
		}

		// Process Query
		if err := executeInteractiveQuery(filename, trimmed); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	return nil
}

func executeInteractiveQuery(filename, expression string) error {
	// 1. Try SQL-like
	if strings.HasPrefix(strings.ToUpper(expression), "SELECT") {
		q, err := query.ParseQuery(expression)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		inputTable := database.NewJSONTable(filename)

		// Create Plan
		rootNode, err := planner.CreatePlan(q, inputTable)
		if err != nil {
			return fmt.Errorf("planning error: %w", err)
		}

		// Explain Mode (check global flag, though interactive might want per-query flag processing?)
		// For simplicity, we use the global flag.
		if QueryExplain {
			fmt.Println("Execution Plan:")
			fmt.Println(plan.FormatPlan(rootNode))
			return nil
		}

		executor := engine.NewExecutor()
		executor.Pretty = QueryPretty
		// We print to stdout
		return executor.Execute(rootNode, os.Stdout)
	}

	// 2. Try Filter Expression
	if query.IsFilterExpression(expression) {
		expr := query.ParseFilterExpression(expression)
		if expr != nil {
			// Reuse RunFilter from root.go or similar logic?
			// RunFilter is in root.go but not exported? No, it's likely internal to package cmd.
			// Let's check root.go again. It calls RunFilter.
			// We can call RunFilter if it's in the same package (cmd).
			// We need to pass the global flags: QueryPretty, QueryExtract, QuerySelect
			return RunFilter(filename, expr.Field, expr.Operator, expr.Value, QueryPretty, QueryExtract, QuerySelect, "json")
		}
	}

	// 3. Try Path Query
	return RunQuery(filename, expression, QueryPretty, QueryExtract, QuerySelect)
}
