package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks by title, description, or tags",
	Long: `Search tasks using a query string that matches against task titles, descriptions, and tags.
The search is case-insensitive and supports partial matches.

Examples:
  task-cli search "bug"
  task-cli search "login" --limit 10
  task-cli search "urgent" --assignee john
  task-cli search "api" --status pending --priority high`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchTasks(cmd, args)
	},
}

func init() {
	searchCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	searchCmd.Flags().StringSliceP("status", "s", []string{}, "Filter by status (pending, in-progress, completed, cancelled, on-hold)")
	searchCmd.Flags().StringSliceP("priority", "p", []string{}, "Filter by priority (low, medium, high, critical)")
	searchCmd.Flags().StringSliceP("tags", "t", []string{}, "Filter by tags")
	searchCmd.Flags().IntP("limit", "l", 20, "Maximum number of results")
	searchCmd.Flags().IntP("offset", "o", 0, "Number of results to skip")
	searchCmd.Flags().String("sort-by", "", "Sort by field (created_at, updated_at, due_date, priority, status)")
	searchCmd.Flags().String("sort-order", "", "Sort order (asc, desc)")
	searchCmd.Flags().Bool("show-completed", false, "Include completed tasks in search")
}

func runSearchTasks(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get search query
	var query string
	if len(args) > 0 {
		query = strings.Join(args, " ")
	}

	if query == "" {
		printError("Search query is required")
		return fmt.Errorf("search query is required")
	}

	// Get flags
	assignee, _ := cmd.Flags().GetString("assignee")
	statusList, _ := cmd.Flags().GetStringSlice("status")
	priorityList, _ := cmd.Flags().GetStringSlice("priority")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sortBy, _ := cmd.Flags().GetString("sort-by")
	sortOrder, _ := cmd.Flags().GetString("sort-order")
	showCompleted, _ := cmd.Flags().GetBool("show-completed")

	// Build filter
	filter := models.TaskFilter{
		Search: query,
	}

	// Add assignee filter
	if assignee != "" {
		filter.Assignee = []string{assignee}
	}

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
	} else if !showCompleted {
		// Exclude completed tasks by default
		filter.Status = []models.TaskStatus{
			models.StatusPending,
			models.StatusInProgress,
			models.StatusOnHold,
		}
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

	// Add tags filter
	if len(tags) > 0 {
		filter.Tags = tags
	}

	// Search tasks
	tasks, total, err := taskService.ListTasks(ctx, filter, sortBy, sortOrder, offset, limit)
	if err != nil {
		printError("Failed to search tasks: %v", err)
		return err
	}

	// Display results
	if len(tasks) == 0 {
		printInfo("No tasks found matching query: %s", query)
		return nil
	}

	// Print search summary
	printSearchSummary(query, len(tasks), total, offset, limit)

	// Display tasks based on output format
	outputFormat := getOutputFormat()
	switch outputFormat {
	case "json":
		return printTasksJSON(tasks)
	case "yaml":
		return printTasksYAML(tasks)
	case "csv":
		return printTasksCSV(tasks)
	default:
		printTasksTable(tasks)
	}

	return nil
}

func printSearchSummary(query string, shown, total, offset, limit int) {
	if isColorEnabled() {
		fmt.Printf("\n🔍 \033[1mSearch Results for:\033[0m \033[36m%s\033[0m\n", query)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Showing \033[33m%d\033[0m of \033[33m%d\033[0m tasks", shown, total)
		if offset > 0 {
			fmt.Printf(" (offset: %d)", offset)
		}
		fmt.Printf("\n\n")
	} else {
		fmt.Printf("\nSearch Results for: %s\n", query)
		fmt.Printf("Showing %d of %d tasks", shown, total)
		if offset > 0 {
			fmt.Printf(" (offset: %d)", offset)
		}
		fmt.Printf("\n\n")
	}
}

func printTasksTable(tasks []*models.Task) {
	if len(tasks) == 0 {
		return
	}

	// Calculate column widths
	maxIDWidth := 8
	maxTitleWidth := 30
	maxStatusWidth := 12
	maxPriorityWidth := 10
	maxAssigneeWidth := 15
	maxDueDateWidth := 16

	for _, task := range tasks {
		if len(task.ID) > maxIDWidth {
			maxIDWidth = len(task.ID)
		}
		if len(task.Title) > maxTitleWidth {
			maxTitleWidth = len(task.Title)
		}
		if len(string(task.Status)) > maxStatusWidth {
			maxStatusWidth = len(string(task.Status))
		}
		if len(string(task.Priority)) > maxPriorityWidth {
			maxPriorityWidth = len(string(task.Priority))
		}
		if len(task.Assignee) > maxAssigneeWidth {
			maxAssigneeWidth = len(task.Assignee)
		}
	}

	// Limit column widths
	if maxTitleWidth > 50 {
		maxTitleWidth = 50
	}
	if maxAssigneeWidth > 20 {
		maxAssigneeWidth = 20
	}

	// Print header
	if isColorEnabled() {
		fmt.Printf("\033[1m%-*s %-*s %-*s %-*s %-*s %-*s\033[0m\n",
			maxIDWidth, "ID",
			maxTitleWidth, "TITLE",
			maxStatusWidth, "STATUS",
			maxPriorityWidth, "PRIORITY",
			maxAssigneeWidth, "ASSIGNEE",
			maxDueDateWidth, "DUE DATE")
	} else {
		fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-*s\n",
			maxIDWidth, "ID",
			maxTitleWidth, "TITLE",
			maxStatusWidth, "STATUS",
			maxPriorityWidth, "PRIORITY",
			maxAssigneeWidth, "ASSIGNEE",
			maxDueDateWidth, "DUE DATE")
	}

	// Print separator
	fmt.Printf("%s %s %s %s %s %s\n",
		strings.Repeat("-", maxIDWidth),
		strings.Repeat("-", maxTitleWidth),
		strings.Repeat("-", maxStatusWidth),
		strings.Repeat("-", maxPriorityWidth),
		strings.Repeat("-", maxAssigneeWidth),
		strings.Repeat("-", maxDueDateWidth))

	// Print tasks
	for _, task := range tasks {
		// Truncate title if too long
		title := task.Title
		if len(title) > maxTitleWidth {
			title = title[:maxTitleWidth-3] + "..."
		}

		// Truncate assignee if too long
		assignee := task.Assignee
		if len(assignee) > maxAssigneeWidth {
			assignee = assignee[:maxAssigneeWidth-3] + "..."
		}

		// Format due date
		dueDate := ""
		if task.DueDate != nil {
			dueDate = task.DueDate.Format(cfg.CLI.DateFormat)
			if len(dueDate) > maxDueDateWidth {
				dueDate = dueDate[:maxDueDateWidth]
			}
		}

		// Color coding
		if isColorEnabled() {
			statusColor := getStatusColor(task.Status)
			priorityColor := getPriorityColor(task.Priority)
			dueDateColor := ""
			if task.IsOverdue() {
				dueDateColor = "\033[31m" // Red
			} else if task.IsDueToday() {
				dueDateColor = "\033[33m" // Yellow
			}

			fmt.Printf("%-*s %-*s %s%-*s\033[0m %s%-*s\033[0m %-*s %s%-*s\033[0m\n",
				maxIDWidth, task.ID,
				maxTitleWidth, title,
				statusColor, maxStatusWidth, task.Status,
				priorityColor, maxPriorityWidth, task.Priority,
				maxAssigneeWidth, assignee,
				dueDateColor, maxDueDateWidth, dueDate)
		} else {
			fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-*s\n",
				maxIDWidth, task.ID,
				maxTitleWidth, title,
				maxStatusWidth, task.Status,
				maxPriorityWidth, task.Priority,
				maxAssigneeWidth, assignee,
				maxDueDateWidth, dueDate)
		}
	}

	fmt.Println()
}
