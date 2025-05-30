package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new task",
	Long: `Create a new task with the specified title and optional parameters.

The title can be provided as an argument or you'll be prompted to enter it.
You can specify additional details like description, priority, assignee, due date, and tags.

Examples:
  task-cli create "Fix login bug"
  task-cli create "Fix login bug" --description "Users can't login with email" --priority high
  task-cli create "Review PR" --assignee john --due "2024-01-15 14:00"
  task-cli create "Setup CI/CD" --tags "devops,automation" --priority medium`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCreateTask(cmd, args)
	},
}

func init() {
	createCmd.Flags().StringP("description", "d", "", "Task description")
	createCmd.Flags().StringP("priority", "p", "", "Task priority (low, medium, high, critical)")
	createCmd.Flags().StringP("assignee", "a", "", "Task assignee")
	createCmd.Flags().StringP("due", "", "", "Due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "Task tags (comma-separated)")
	createCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for task creation")
}

func runCreateTask(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get task details
	req, err := getTaskCreateRequest(cmd, args)
	if err != nil {
		printError("Failed to get task details: %v", err)
		return err
	}

	// Create task
	task, err := taskService.CreateTask(ctx, req, cfg.CLI.DefaultUser)
	if err != nil {
		printError("Failed to create task: %v", err)
		return err
	}

	// Print success message
	printSuccess("Task created successfully!")
	printTaskDetails(task)

	return nil
}

func getTaskCreateRequest(cmd *cobra.Command, args []string) (models.TaskCreateRequest, error) {
	var req models.TaskCreateRequest

	// Get interactive flag
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Get title
	if len(args) > 0 {
		req.Title = strings.Join(args, " ")
	} else if interactive {
		req.Title = promptForInput("Task title", true)
	} else {
		return req, fmt.Errorf("task title is required")
	}

	// Get description
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		req.Description = desc
	} else if interactive {
		req.Description = promptForInput("Description (optional)", false)
	}

	// Get priority
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		if !models.ValidatePriority(priority) {
			return req, fmt.Errorf("invalid priority: %s (valid: low, medium, high, critical)", priority)
		}
		req.Priority = models.TaskPriority(priority)
	} else if interactive {
		req.Priority = promptForPriority()
	} else {
		req.Priority = models.TaskPriority(cfg.Defaults.Priority)
	}

	// Get assignee
	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		req.Assignee = assignee
	} else if interactive {
		req.Assignee = promptForInput("Assignee (optional)", false)
	}

	// Get due date
	if dueStr, _ := cmd.Flags().GetString("due"); dueStr != "" {
		dueDate, err := parseDueDate(dueStr)
		if err != nil {
			return req, fmt.Errorf("invalid due date format: %v", err)
		}
		req.DueDate = dueDate
	} else if interactive {
		if dueDateStr := promptForInput("Due date (YYYY-MM-DD HH:MM, optional)", false); dueDateStr != "" {
			dueDate, err := parseDueDate(dueDateStr)
			if err != nil {
				return req, fmt.Errorf("invalid due date format: %v", err)
			}
			req.DueDate = dueDate
		}
	}

	// Get tags
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		req.Tags = tags
	} else if interactive {
		if tagsStr := promptForInput("Tags (comma-separated, optional)", false); tagsStr != "" {
			req.Tags = strings.Split(tagsStr, ",")
			for i, tag := range req.Tags {
				req.Tags[i] = strings.TrimSpace(tag)
			}
		}
	}

	return req, nil
}

func parseDueDate(dateStr string) (*time.Time, error) {
	// Try different date formats
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01-02-2006",
		"01/02/2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return &date, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s (expected format: YYYY-MM-DD HH:MM)", dateStr)
}

func promptForInput(prompt string, required bool) string {
	for {
		fmt.Printf("%s: ", prompt)
		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(input)
		if input != "" || !required {
			return input
		}

		printWarning("This field is required")
	}
}

func promptForPriority() models.TaskPriority {
	priorities := models.GetAllPriorities()
	
	fmt.Println("Available priorities:")
	for i, priority := range priorities {
		fmt.Printf("  %d. %s\n", i+1, priority)
	}

	for {
		fmt.Printf("Select priority (1-%d) [default: %s]: ", len(priorities), cfg.Defaults.Priority)
		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(input)
		if input == "" {
			return models.TaskPriority(cfg.Defaults.Priority)
		}

		// Try to parse as number
		if input == "1" {
			return models.PriorityLow
		} else if input == "2" {
			return models.PriorityMedium
		} else if input == "3" {
			return models.PriorityHigh
		} else if input == "4" {
			return models.PriorityCritical
		}

		// Try to parse as string
		if models.ValidatePriority(input) {
			return models.TaskPriority(input)
		}

		printWarning("Invalid priority. Please select 1-4 or enter a valid priority name.")
	}
}

func printTaskDetails(task *models.Task) {
	if !isColorEnabled() {
		printTaskDetailsPlain(task)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("\033[1m📋 Task Details\033[0m\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\033[36mID:\033[0m          %s\n", task.ID)
	fmt.Printf("\033[36mTitle:\033[0m       %s\n", task.Title)
	
	if task.Description != "" {
		fmt.Printf("\033[36mDescription:\033[0m %s\n", task.Description)
	}
	
	// Color-coded status
	statusColor := getStatusColor(task.Status)
	fmt.Printf("\033[36mStatus:\033[0m      %s%s\033[0m\n", statusColor, task.Status)
	
	// Color-coded priority
	priorityColor := getPriorityColor(task.Priority)
	fmt.Printf("\033[36mPriority:\033[0m    %s%s\033[0m\n", priorityColor, task.Priority)
	
	if task.Assignee != "" {
		fmt.Printf("\033[36mAssignee:\033[0m    %s\n", task.Assignee)
	}
	
	fmt.Printf("\033[36mCreator:\033[0m     %s\n", task.Creator)
	fmt.Printf("\033[36mCreated:\033[0m     %s\n", task.CreatedAt.Format(cfg.CLI.DateFormat))
	
	if task.DueDate != nil {
		dueColor := "\033[0m"
		if task.IsOverdue() {
			dueColor = "\033[31m" // Red for overdue
		} else if task.IsDueToday() {
			dueColor = "\033[33m" // Yellow for due today
		}
		fmt.Printf("\033[36mDue Date:\033[0m    %s%s\033[0m\n", dueColor, task.DueDate.Format(cfg.CLI.DateFormat))
	}
	
	if len(task.Tags) > 0 {
		fmt.Printf("\033[36mTags:\033[0m        %s\n", strings.Join(task.Tags, ", "))
	}
	
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func printTaskDetailsPlain(task *models.Task) {
	fmt.Printf("\nTask Details:\n")
	fmt.Printf("ID:          %s\n", task.ID)
	fmt.Printf("Title:       %s\n", task.Title)
	
	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}
	
	fmt.Printf("Status:      %s\n", task.Status)
	fmt.Printf("Priority:    %s\n", task.Priority)
	
	if task.Assignee != "" {
		fmt.Printf("Assignee:    %s\n", task.Assignee)
	}
	
	fmt.Printf("Creator:     %s\n", task.Creator)
	fmt.Printf("Created:     %s\n", task.CreatedAt.Format(cfg.CLI.DateFormat))
	
	if task.DueDate != nil {
		fmt.Printf("Due Date:    %s\n", task.DueDate.Format(cfg.CLI.DateFormat))
	}
	
	if len(task.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(task.Tags, ", "))
	}
}

func getStatusColor(status models.TaskStatus) string {
	switch status {
	case models.StatusPending:
		return "\033[33m" // Yellow
	case models.StatusInProgress:
		return "\033[34m" // Blue
	case models.StatusCompleted:
		return "\033[32m" // Green
	case models.StatusCancelled:
		return "\033[31m" // Red
	case models.StatusOnHold:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Default
	}
}

func getPriorityColor(priority models.TaskPriority) string {
	switch priority {
	case models.PriorityLow:
		return "\033[32m" // Green
	case models.PriorityMedium:
		return "\033[33m" // Yellow
	case models.PriorityHigh:
		return "\033[35m" // Magenta
	case models.PriorityCritical:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Default
	}
}
