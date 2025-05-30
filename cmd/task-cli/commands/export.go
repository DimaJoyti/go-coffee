package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var exportCmd = &cobra.Command{
	Use:   "export [filename]",
	Short: "Export tasks to various formats",
	Long: `Export tasks to JSON, CSV, or YAML format.
If no filename is provided, the export will be written to stdout.
The format is determined by the file extension or the --format flag.

Supported formats:
  • json - JSON format
  • csv  - Comma-separated values
  • yaml - YAML format

Examples:
  task-cli export tasks.json
  task-cli export tasks.csv --status pending
  task-cli export --format yaml --assignee john
  task-cli export backup.json --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportTasks(cmd, args)
	},
}

func init() {
	exportCmd.Flags().StringP("format", "f", "", "Export format (json, csv, yaml)")
	exportCmd.Flags().StringSliceP("status", "s", []string{}, "Filter by status")
	exportCmd.Flags().StringSliceP("priority", "p", []string{}, "Filter by priority")
	exportCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	exportCmd.Flags().StringSliceP("tags", "t", []string{}, "Filter by tags")
	exportCmd.Flags().String("due-before", "", "Filter tasks due before date (YYYY-MM-DD)")
	exportCmd.Flags().String("due-after", "", "Filter tasks due after date (YYYY-MM-DD)")
	exportCmd.Flags().Bool("all", false, "Export all tasks (ignore other filters)")
	exportCmd.Flags().Bool("pretty", false, "Pretty print JSON output")
}

func runExportTasks(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	format, _ := cmd.Flags().GetString("format")
	statusList, _ := cmd.Flags().GetStringSlice("status")
	priorityList, _ := cmd.Flags().GetStringSlice("priority")
	assignee, _ := cmd.Flags().GetString("assignee")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	dueBefore, _ := cmd.Flags().GetString("due-before")
	dueAfter, _ := cmd.Flags().GetString("due-after")
	exportAll, _ := cmd.Flags().GetBool("all")
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Determine filename and format
	var filename string
	if len(args) > 0 {
		filename = args[0]
		if format == "" {
			// Determine format from file extension
			ext := strings.ToLower(filepath.Ext(filename))
			switch ext {
			case ".json":
				format = "json"
			case ".csv":
				format = "csv"
			case ".yaml", ".yml":
				format = "yaml"
			default:
				format = "json" // Default to JSON
			}
		}
	} else {
		// No filename provided, output to stdout
		if format == "" {
			format = "json" // Default format
		}
	}

	// Build filter
	filter := models.TaskFilter{}

	if !exportAll {
		// Add status filter
		if len(statusList) > 0 {
			var statuses []models.TaskStatus
			for _, s := range statusList {
				if models.ValidateStatus(s) {
					statuses = append(statuses, models.TaskStatus(s))
				} else {
					printWarning("Invalid status: %s", s)
				}
			}
			filter.Status = statuses
		}

		// Add priority filter
		if len(priorityList) > 0 {
			var priorities []models.TaskPriority
			for _, p := range priorityList {
				if models.ValidatePriority(p) {
					priorities = append(priorities, models.TaskPriority(p))
				} else {
					printWarning("Invalid priority: %s", p)
				}
			}
			filter.Priority = priorities
		}

		// Add assignee filter
		if assignee != "" {
			filter.Assignee = []string{assignee}
		}

		// Add tags filter
		if len(tags) > 0 {
			filter.Tags = tags
		}

		// Add date filters
		if dueBefore != "" {
			date, err := time.Parse("2006-01-02", dueBefore)
			if err != nil {
				printError("Invalid due-before date format: %s (expected YYYY-MM-DD)", dueBefore)
				return err
			}
			filter.DueBefore = &date
		}

		if dueAfter != "" {
			date, err := time.Parse("2006-01-02", dueAfter)
			if err != nil {
				printError("Invalid due-after date format: %s (expected YYYY-MM-DD)", dueAfter)
				return err
			}
			filter.DueAfter = &date
		}
	}

	// Get tasks
	tasks, total, err := taskService.ListTasks(ctx, filter, "", "", 0, 10000) // Get all matching tasks
	if err != nil {
		printError("Failed to get tasks for export: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks found matching the specified criteria")
		return nil
	}

	// Export tasks
	var data []byte
	switch format {
	case "csv":
		data, err = exportToCSV(tasks)
	case "yaml":
		data, err = exportToYAML(tasks)
	default: // json
		data, err = exportToJSON(tasks, pretty)
	}

	if err != nil {
		printError("Failed to export tasks: %v", err)
		return err
	}

	// Write to file or stdout
	if filename != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(filename)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				printError("Failed to create directory: %v", err)
				return err
			}
		}

		// Write to file
		if err := os.WriteFile(filename, data, 0644); err != nil {
			printError("Failed to write to file: %v", err)
			return err
		}

		printSuccess("Exported %d tasks to %s", len(tasks), filename)
		printInfo("Format: %s", format)
		if total > len(tasks) {
			printInfo("Note: %d tasks matched filters, %d exported", total, len(tasks))
		}
	} else {
		// Write to stdout
		fmt.Print(string(data))
	}

	return nil
}

func exportToJSON(tasks []*models.Task, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(tasks, "", "  ")
	}
	return json.Marshal(tasks)
}

func exportToYAML(tasks []*models.Task) ([]byte, error) {
	return yaml.Marshal(tasks)
}

func exportToCSV(tasks []*models.Task) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"ID", "Title", "Description", "Status", "Priority",
		"Assignee", "Creator", "Tags", "Created At", "Updated At",
		"Due Date", "Completed At",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write tasks
	for _, task := range tasks {
		record := []string{
			task.ID,
			task.Title,
			task.Description,
			string(task.Status),
			string(task.Priority),
			task.Assignee,
			task.Creator,
			strings.Join(task.Tags, ";"),
			task.CreatedAt.Format(time.RFC3339),
			task.UpdatedAt.Format(time.RFC3339),
		}

		// Add due date
		if task.DueDate != nil {
			record = append(record, task.DueDate.Format(time.RFC3339))
		} else {
			record = append(record, "")
		}

		// Add completed date
		if task.CompletedAt != nil {
			record = append(record, task.CompletedAt.Format(time.RFC3339))
		} else {
			record = append(record, "")
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(buf.String()), nil
}

// Helper functions for other commands to use

func printTasksJSON(tasks []*models.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printTasksYAML(tasks []*models.Task) error {
	data, err := yaml.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to YAML: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printTasksCSV(tasks []*models.Task) error {
	data, err := exportToCSV(tasks)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
