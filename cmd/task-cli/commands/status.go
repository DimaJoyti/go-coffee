package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [task-id] [new-status]",
	Short: "Change the status of a task",
	Long: `Change the status of a task to one of the valid statuses:
  • pending
  • in-progress
  • completed
  • cancelled
  • on-hold

Examples:
  task-cli status 123 completed
  task-cli status abc123 in-progress
  task-cli status --interactive
  task-cli status --bulk --status pending --new-status in-progress`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangeStatus(cmd, args)
	},
}

func init() {
	statusCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for status change")
	statusCmd.Flags().Bool("bulk", false, "Bulk status change mode")
	statusCmd.Flags().String("status", "", "Current status filter for bulk operations")
	statusCmd.Flags().String("new-status", "", "New status for bulk operations")
	statusCmd.Flags().StringP("assignee", "a", "", "Filter by assignee for bulk operations")
	statusCmd.Flags().StringSliceP("priority", "p", []string{}, "Filter by priority for bulk operations")
	statusCmd.Flags().StringSliceP("tags", "t", []string{}, "Filter by tags for bulk operations")
	statusCmd.Flags().Bool("confirm", false, "Skip confirmation for bulk operations")
}

func runChangeStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Check for bulk mode
	bulk, _ := cmd.Flags().GetBool("bulk")
	if bulk {
		return runBulkStatusChange(cmd, ctx)
	}

	// Check for interactive mode
	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		return runInteractiveStatusChange(cmd, ctx)
	}

	// Regular mode - require task ID and new status
	if len(args) < 2 {
		printError("Task ID and new status are required")
		printInfo("Usage: task-cli status [task-id] [new-status]")
		printInfo("Valid statuses: %s", strings.Join(getValidStatusStrings(), ", "))
		return fmt.Errorf("insufficient arguments")
	}

	taskID := args[0]
	newStatus := args[1]

	// Validate status
	if !models.ValidateStatus(newStatus) {
		printError("Invalid status: %s", newStatus)
		printInfo("Valid statuses: %s", strings.Join(getValidStatusStrings(), ", "))
		return fmt.Errorf("invalid status")
	}

	// Change status
	task, err := taskService.ChangeTaskStatus(ctx, taskID, models.TaskStatus(newStatus))
	if err != nil {
		printError("Failed to change task status: %v", err)
		return err
	}

	// Print success message
	printSuccess("Task status changed successfully!")
	printTaskDetails(task)

	return nil
}

func runBulkStatusChange(cmd *cobra.Command, ctx context.Context) error {
	// Get bulk operation parameters
	currentStatus, _ := cmd.Flags().GetString("status")
	newStatus, _ := cmd.Flags().GetString("new-status")
	assignee, _ := cmd.Flags().GetString("assignee")
	priorities, _ := cmd.Flags().GetStringSlice("priority")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	skipConfirm, _ := cmd.Flags().GetBool("confirm")

	// Validate required parameters
	if newStatus == "" {
		printError("New status is required for bulk operations")
		return fmt.Errorf("new status required")
	}

	if !models.ValidateStatus(newStatus) {
		printError("Invalid new status: %s", newStatus)
		printInfo("Valid statuses: %s", strings.Join(getValidStatusStrings(), ", "))
		return fmt.Errorf("invalid status")
	}

	// Build filter
	filter := models.TaskFilter{}

	if currentStatus != "" {
		if !models.ValidateStatus(currentStatus) {
			printError("Invalid current status: %s", currentStatus)
			return fmt.Errorf("invalid status")
		}
		filter.Status = []models.TaskStatus{models.TaskStatus(currentStatus)}
	}

	if assignee != "" {
		filter.Assignee = []string{assignee}
	}

	if len(priorities) > 0 {
		var validPriorities []models.TaskPriority
		for _, p := range priorities {
			if models.ValidatePriority(p) {
				validPriorities = append(validPriorities, models.TaskPriority(p))
			} else {
				printWarning("Invalid priority: %s", p)
			}
		}
		filter.Priority = validPriorities
	}

	if len(tags) > 0 {
		filter.Tags = tags
	}

	// Get matching tasks
	tasks, _, err := taskService.ListTasks(ctx, filter, "", "", 0, 1000)
	if err != nil {
		printError("Failed to get tasks for bulk operation: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks found matching the specified criteria")
		return nil
	}

	// Show preview and confirm
	if !skipConfirm {
		printInfo("Found %d tasks matching criteria:", len(tasks))
		for i, task := range tasks {
			if i < 5 { // Show first 5 tasks
				fmt.Printf("  - %s: %s (current: %s)\n", task.ID, task.Title, task.Status)
			}
		}
		if len(tasks) > 5 {
			fmt.Printf("  ... and %d more\n", len(tasks)-5)
		}

		fmt.Printf("\nChange status from various to '%s'? (y/N): ", newStatus)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			printInfo("Operation cancelled")
			return nil
		}
	}

	// Perform bulk update
	updateReq := models.TaskUpdateRequest{
		Status: (*models.TaskStatus)(&newStatus),
	}

	updated, err := taskService.BulkUpdateTasks(ctx, filter, updateReq)
	if err != nil {
		printError("Failed to perform bulk status change: %v", err)
		return err
	}

	printSuccess("Successfully updated %d tasks to status '%s'", updated, newStatus)
	return nil
}

func runInteractiveStatusChange(cmd *cobra.Command, ctx context.Context) error {
	// Get task ID
	fmt.Print("Enter task ID: ")
	var taskID string
	fmt.Scanln(&taskID)

	if taskID == "" {
		printError("Task ID is required")
		return fmt.Errorf("task ID required")
	}

	// Get current task
	task, err := taskService.GetTask(ctx, taskID)
	if err != nil {
		printError("Failed to get task: %v", err)
		return err
	}

	// Show current task info
	fmt.Printf("\nCurrent task: %s\n", task.Title)
	fmt.Printf("Current status: %s\n\n", task.Status)

	// Show available statuses
	statuses := models.GetAllStatuses()
	fmt.Println("Available statuses:")
	for i, status := range statuses {
		marker := " "
		if status == task.Status {
			marker = ">"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, status)
	}

	// Get new status
	fmt.Printf("\nSelect new status (1-%d): ", len(statuses))
	var choice string
	fmt.Scanln(&choice)

	// Parse choice
	var newStatus models.TaskStatus
	switch choice {
	case "1":
		newStatus = models.StatusPending
	case "2":
		newStatus = models.StatusInProgress
	case "3":
		newStatus = models.StatusCompleted
	case "4":
		newStatus = models.StatusCancelled
	case "5":
		newStatus = models.StatusOnHold
	default:
		// Try to parse as status name
		if models.ValidateStatus(choice) {
			newStatus = models.TaskStatus(choice)
		} else {
			printError("Invalid choice: %s", choice)
			return fmt.Errorf("invalid choice")
		}
	}

	// Confirm change
	if newStatus == task.Status {
		printInfo("Status is already '%s'", newStatus)
		return nil
	}

	fmt.Printf("Change status from '%s' to '%s'? (y/N): ", task.Status, newStatus)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		printInfo("Operation cancelled")
		return nil
	}

	// Change status
	updatedTask, err := taskService.ChangeTaskStatus(ctx, taskID, newStatus)
	if err != nil {
		printError("Failed to change task status: %v", err)
		return err
	}

	printSuccess("Task status changed successfully!")
	printTaskDetails(updatedTask)

	return nil
}

func getValidStatusStrings() []string {
	statuses := models.GetAllStatuses()
	var strings []string
	for _, status := range statuses {
		strings = append(strings, string(status))
	}
	return strings
}
