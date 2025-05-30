package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks with filtering and sorting options",
	Long: `List tasks with powerful filtering, sorting, and pagination options.

You can filter tasks by status, priority, assignee, creator, tags, and due dates.
Results can be sorted by various fields and displayed in different formats.

Examples:
  task-cli list
  task-cli list --status pending,in-progress
  task-cli list --assignee john --priority high
  task-cli list --due-before "2024-01-15"
  task-cli list --tags "bug,urgent" --sort-by priority --sort-order desc
  task-cli list --output json
  task-cli list --search "login"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runListTasks(cmd, args)
	},
}

func init() {
	listCmd.Flags().StringSlice("status", []string{}, "Filter by status (pending,in-progress,completed,cancelled,on-hold)")
	listCmd.Flags().StringSlice("priority", []string{}, "Filter by priority (low,medium,high,critical)")
	listCmd.Flags().StringSlice("assignee", []string{}, "Filter by assignee")
	listCmd.Flags().StringSlice("creator", []string{}, "Filter by creator")
	listCmd.Flags().StringSlice("tags", []string{}, "Filter by tags")
	listCmd.Flags().String("due-before", "", "Filter tasks due before date (YYYY-MM-DD)")
	listCmd.Flags().String("due-after", "", "Filter tasks due after date (YYYY-MM-DD)")
	listCmd.Flags().String("search", "", "Search in title and description")
	listCmd.Flags().String("sort-by", "", "Sort by field (title,status,priority,assignee,creator,created_at,updated_at,due_date)")
	listCmd.Flags().String("sort-order", "", "Sort order (asc,desc)")
	listCmd.Flags().Int("limit", 0, "Limit number of results")
	listCmd.Flags().Int("offset", 0, "Offset for pagination")
	listCmd.Flags().Bool("overdue", false, "Show only overdue tasks")
	listCmd.Flags().Bool("due-today", false, "Show only tasks due today")
	listCmd.Flags().Bool("due-this-week", false, "Show only tasks due this week")
	listCmd.Flags().Bool("my-tasks", false, "Show only tasks assigned to current user")
}

func runListTasks(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Build filter
	filter, err := buildTaskFilter(cmd)
	if err != nil {
		printError("Invalid filter: %v", err)
		return err
	}

	// Get sorting options
	sortBy, _ := cmd.Flags().GetString("sort-by")
	sortOrder, _ := cmd.Flags().GetString("sort-order")

	// Get pagination options
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")

	if limit == 0 {
		limit = cfg.CLI.PageSize
	}

	// Handle special filters
	var tasks []*models.Task
	var total int

	if overdue, _ := cmd.Flags().GetBool("overdue"); overdue {
		tasks, err = taskService.GetOverdueTasks(ctx)
		total = len(tasks)
	} else if dueToday, _ := cmd.Flags().GetBool("due-today"); dueToday {
		tasks, err = taskService.GetTasksDueToday(ctx)
		total = len(tasks)
	} else if dueThisWeek, _ := cmd.Flags().GetBool("due-this-week"); dueThisWeek {
		tasks, err = taskService.GetTasksDueThisWeek(ctx)
		total = len(tasks)
	} else if myTasks, _ := cmd.Flags().GetBool("my-tasks"); myTasks {
		tasks, err = taskService.GetTasksByAssignee(ctx, cfg.CLI.DefaultUser)
		total = len(tasks)
	} else {
		tasks, total, err = taskService.ListTasks(ctx, filter, sortBy, sortOrder, offset, limit)
	}

	if err != nil {
		printError("Failed to list tasks: %v", err)
		return err
	}

	// Display results
	return displayTasks(tasks, total, offset, limit)
}

func buildTaskFilter(cmd *cobra.Command) (models.TaskFilter, error) {
	var filter models.TaskFilter

	// Status filter
	if statusList, _ := cmd.Flags().GetStringSlice("status"); len(statusList) > 0 {
		for _, status := range statusList {
			if !models.ValidateStatus(status) {
				return filter, fmt.Errorf("invalid status: %s", status)
			}
			filter.Status = append(filter.Status, models.TaskStatus(status))
		}
	}

	// Priority filter
	if priorityList, _ := cmd.Flags().GetStringSlice("priority"); len(priorityList) > 0 {
		for _, priority := range priorityList {
			if !models.ValidatePriority(priority) {
				return filter, fmt.Errorf("invalid priority: %s", priority)
			}
			filter.Priority = append(filter.Priority, models.TaskPriority(priority))
		}
	}

	// Assignee filter
	if assigneeList, _ := cmd.Flags().GetStringSlice("assignee"); len(assigneeList) > 0 {
		filter.Assignee = assigneeList
	}

	// Creator filter
	if creatorList, _ := cmd.Flags().GetStringSlice("creator"); len(creatorList) > 0 {
		filter.Creator = creatorList
	}

	// Tags filter
	if tagsList, _ := cmd.Flags().GetStringSlice("tags"); len(tagsList) > 0 {
		filter.Tags = tagsList
	}

	// Due date filters
	if dueBefore, _ := cmd.Flags().GetString("due-before"); dueBefore != "" {
		date, err := time.Parse("2006-01-02", dueBefore)
		if err != nil {
			return filter, fmt.Errorf("invalid due-before date format: %s (expected YYYY-MM-DD)", dueBefore)
		}
		filter.DueBefore = &date
	}

	if dueAfter, _ := cmd.Flags().GetString("due-after"); dueAfter != "" {
		date, err := time.Parse("2006-01-02", dueAfter)
		if err != nil {
			return filter, fmt.Errorf("invalid due-after date format: %s (expected YYYY-MM-DD)", dueAfter)
		}
		filter.DueAfter = &date
	}

	// Search filter
	if search, _ := cmd.Flags().GetString("search"); search != "" {
		filter.Search = search
	}

	return filter, nil
}

func displayTasks(tasks []*models.Task, total, offset, limit int) error {
	if len(tasks) == 0 {
		printInfo("No tasks found")
		return nil
	}

	format := getOutputFormat()

	switch format {
	case "json":
		return displayTasksJSON(tasks)
	case "yaml":
		return displayTasksYAML(tasks)
	case "csv":
		return displayTasksCSV(tasks)
	default:
		return displayTasksTable(tasks, total, offset, limit)
	}
}

func displayTasksTable(tasks []*models.Task, total, offset, limit int) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	if isColorEnabled() {
		fmt.Fprintf(w, "\033[1m%s\t%s\t%s\t%s\t%s\t%s\t%s\033[0m\n",
			"ID", "TITLE", "STATUS", "PRIORITY", "ASSIGNEE", "DUE DATE", "CREATED")
	} else {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			"ID", "TITLE", "STATUS", "PRIORITY", "ASSIGNEE", "DUE DATE", "CREATED")
	}

	// Print separator
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		strings.Repeat("-", 8), strings.Repeat("-", 30), strings.Repeat("-", 12),
		strings.Repeat("-", 10), strings.Repeat("-", 15), strings.Repeat("-", 12), strings.Repeat("-", 12))

	// Print tasks
	for _, task := range tasks {
		id := task.ID
		if len(id) > 8 {
			id = id[:8]
		}

		title := task.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}

		assignee := task.Assignee
		if assignee == "" {
			assignee = "-"
		}
		if len(assignee) > 15 {
			assignee = assignee[:12] + "..."
		}

		dueDate := "-"
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02")
		}

		createdDate := task.CreatedAt.Format("2006-01-02")

		if isColorEnabled() {
			statusColor := getStatusColor(task.Status)
			priorityColor := getPriorityColor(task.Priority)
			dueDateColor := getDueDateColor(task)

			fmt.Fprintf(w, "%s\t%s\t%s%s\033[0m\t%s%s\033[0m\t%s\t%s%s\033[0m\t%s\n",
				id, title, statusColor, task.Status, priorityColor, task.Priority,
				assignee, dueDateColor, dueDate, createdDate)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				id, title, task.Status, task.Priority, assignee, dueDate, createdDate)
		}
	}

	w.Flush()

	// Print pagination info
	if total > len(tasks) {
		fmt.Printf("\nShowing %d-%d of %d tasks", offset+1, offset+len(tasks), total)
		if offset+limit < total {
			fmt.Printf(" (use --offset %d to see more)", offset+limit)
		}
		fmt.Println()
	} else {
		fmt.Printf("\nTotal: %d tasks\n", total)
	}

	return nil
}

func displayTasksJSON(tasks []*models.Task) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

func displayTasksYAML(tasks []*models.Task) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(tasks)
}

func displayTasksCSV(tasks []*models.Task) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Title", "Description", "Status", "Priority", "Assignee", "Creator", "Tags", "Created", "Updated", "Due Date", "Completed"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write tasks
	for _, task := range tasks {
		dueDate := ""
		if task.DueDate != nil {
			dueDate = task.DueDate.Format(time.RFC3339)
		}

		completedAt := ""
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt.Format(time.RFC3339)
		}

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
			dueDate,
			completedAt,
		}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func getDueDateColor(task *models.Task) string {
	if task.DueDate == nil {
		return "\033[0m"
	}

	if task.IsOverdue() {
		return "\033[31m" // Red
	}

	if task.IsDueToday() {
		return "\033[33m" // Yellow
	}

	return "\033[0m" // Default
}
