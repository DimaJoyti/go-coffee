package commands

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign [task-id] [assignee]",
	Short: "Assign a task to a user",
	Long: `Assign a task to a specific user.

You can assign a single task to a user by providing the task ID and assignee name.
Use the --bulk flag to assign multiple tasks at once.

Examples:
  task-cli assign 123 john
  task-cli assign 123 john.doe@company.com
  task-cli assign --bulk --status pending --assignee john`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAssignTask(cmd, args)
	},
}

func init() {
	assignCmd.Flags().Bool("bulk", false, "Bulk assign tasks based on filters")
	assignCmd.Flags().String("status", "", "Filter tasks by status for bulk assignment")
	assignCmd.Flags().String("priority", "", "Filter tasks by priority for bulk assignment")
	assignCmd.Flags().String("creator", "", "Filter tasks by creator for bulk assignment")
	assignCmd.Flags().StringSlice("tags", []string{}, "Filter tasks by tags for bulk assignment")
	assignCmd.Flags().String("assignee", "", "New assignee for bulk assignment")
	assignCmd.Flags().Bool("unassign", false, "Remove assignee from task(s)")
}

func runAssignTask(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	bulk, _ := cmd.Flags().GetBool("bulk")
	unassign, _ := cmd.Flags().GetBool("unassign")

	if bulk {
		return runBulkAssign(ctx, cmd)
	}

	// Single task assignment
	if len(args) < 1 {
		printError("Task ID is required")
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]
	var assignee string

	if unassign {
		assignee = ""
	} else {
		if len(args) < 2 {
			assigneeFlag, _ := cmd.Flags().GetString("assignee")
			if assigneeFlag == "" {
				printError("Assignee is required (use --unassign to remove assignee)")
				return fmt.Errorf("assignee is required")
			}
			assignee = assigneeFlag
		} else {
			assignee = args[1]
		}
	}

	// Get current task to show before/after
	currentTask, err := taskService.GetTask(ctx, taskID)
	if err != nil {
		printError("Failed to get task: %v", err)
		return err
	}

	// Assign task
	updatedTask, err := taskService.AssignTask(ctx, taskID, assignee)
	if err != nil {
		printError("Failed to assign task: %v", err)
		return err
	}

	// Print success message
	if unassign || assignee == "" {
		printSuccess("Task unassigned successfully!")
		printInfo("Task '%s' is no longer assigned to anyone", updatedTask.Title)
	} else {
		printSuccess("Task assigned successfully!")
		if currentTask.Assignee != "" && currentTask.Assignee != assignee {
			printInfo("Task '%s' reassigned from '%s' to '%s'", updatedTask.Title, currentTask.Assignee, assignee)
		} else {
			printInfo("Task '%s' assigned to '%s'", updatedTask.Title, assignee)
		}
	}

	return nil
}

func runBulkAssign(ctx context.Context, cmd *cobra.Command) error {
	// Build filter for bulk assignment
	filter, err := buildBulkAssignFilter(cmd)
	if err != nil {
		printError("Invalid filter: %v", err)
		return err
	}

	assignee, _ := cmd.Flags().GetString("assignee")
	unassign, _ := cmd.Flags().GetBool("unassign")

	if !unassign && assignee == "" {
		printError("Assignee is required for bulk assignment (use --unassign to remove assignees)")
		return fmt.Errorf("assignee is required")
	}

	if unassign {
		assignee = ""
	}

	// Get tasks matching the filter
	tasks, _, err := taskService.ListTasks(ctx, filter, "", "", 0, 10000)
	if err != nil {
		printError("Failed to get tasks: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks found matching the filter")
		return nil
	}

	// Show what will be updated
	fmt.Printf("The following %d task(s) will be ", len(tasks))
	if unassign {
		fmt.Printf("unassigned:\n")
	} else {
		fmt.Printf("assigned to '%s':\n", assignee)
	}

	for _, task := range tasks {
		currentAssignee := task.Assignee
		if currentAssignee == "" {
			currentAssignee = "(unassigned)"
		}
		fmt.Printf("- %s: %s [currently: %s]\n", task.ID[:8], task.Title, currentAssignee)
	}

	// Confirm bulk assignment
	fmt.Printf("\nProceed with bulk assignment? (y/N): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "yes" {
		printInfo("Bulk assignment cancelled")
		return nil
	}

	// Perform bulk assignment
	assigned := 0
	for _, task := range tasks {
		_, err := taskService.AssignTask(ctx, task.ID, assignee)
		if err != nil {
			printError("Failed to assign task %s: %v", task.ID, err)
			continue
		}
		assigned++
	}

	if unassign {
		printSuccess("Successfully unassigned %d task(s)", assigned)
	} else {
		printSuccess("Successfully assigned %d task(s) to '%s'", assigned, assignee)
	}

	if assigned < len(tasks) {
		printWarning("Failed to assign %d task(s)", len(tasks)-assigned)
	}

	return nil
}

func buildBulkAssignFilter(cmd *cobra.Command) (models.TaskFilter, error) {
	var filter models.TaskFilter

	// Status filter
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		if !models.ValidateStatus(status) {
			return filter, fmt.Errorf("invalid status: %s", status)
		}
		filter.Status = []models.TaskStatus{models.TaskStatus(status)}
	}

	// Priority filter
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		if !models.ValidatePriority(priority) {
			return filter, fmt.Errorf("invalid priority: %s", priority)
		}
		filter.Priority = []models.TaskPriority{models.TaskPriority(priority)}
	}

	// Creator filter
	if creator, _ := cmd.Flags().GetString("creator"); creator != "" {
		filter.Creator = []string{creator}
	}

	// Tags filter
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	return filter, nil
}
