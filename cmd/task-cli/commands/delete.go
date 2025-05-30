package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [task-id...]",
	Short: "Delete one or more tasks",
	Long: `Delete one or more tasks by their IDs.

You can delete a single task or multiple tasks at once.
Use the --force flag to skip confirmation prompts.

Examples:
  task-cli delete 123
  task-cli delete 123 456 789
  task-cli delete 123 --force
  task-cli delete --all --force`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeleteTask(cmd, args)
	},
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().Bool("all", false, "Delete all tasks (requires --force)")
	deleteCmd.Flags().String("status", "", "Delete all tasks with specific status")
	deleteCmd.Flags().String("assignee", "", "Delete all tasks assigned to specific user")
}

func runDeleteTask(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	force, _ := cmd.Flags().GetBool("force")
	deleteAll, _ := cmd.Flags().GetBool("all")
	status, _ := cmd.Flags().GetString("status")
	assignee, _ := cmd.Flags().GetString("assignee")

	// Handle bulk delete operations
	if deleteAll {
		return runDeleteAllTasks(ctx, force)
	}

	if status != "" {
		return runDeleteTasksByStatus(ctx, status, force)
	}

	if assignee != "" {
		return runDeleteTasksByAssignee(ctx, assignee, force)
	}

	// Handle individual task deletion
	if len(args) == 0 {
		printError("No task IDs provided")
		return fmt.Errorf("task ID is required")
	}

	return runDeleteTasksByIDs(ctx, args, force)
}

func runDeleteTasksByIDs(ctx context.Context, taskIDs []string, force bool) error {
	// Get tasks to show what will be deleted
	var tasksToDelete []string
	var validTasks []string

	for _, taskID := range taskIDs {
		task, err := taskService.GetTask(ctx, taskID)
		if err != nil {
			printWarning("Task %s not found, skipping", taskID)
			continue
		}
		tasksToDelete = append(tasksToDelete, fmt.Sprintf("- %s: %s", task.ID[:8], task.Title))
		validTasks = append(validTasks, taskID)
	}

	if len(validTasks) == 0 {
		printError("No valid tasks found to delete")
		return fmt.Errorf("no valid tasks found")
	}

	// Show confirmation
	if !force {
		fmt.Printf("The following tasks will be deleted:\n")
		for _, taskInfo := range tasksToDelete {
			fmt.Println(taskInfo)
		}
		fmt.Printf("\nAre you sure you want to delete %d task(s)? (y/N): ", len(validTasks))
		
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		
		if response != "y" && response != "yes" {
			printInfo("Deletion cancelled")
			return nil
		}
	}

	// Delete tasks
	deleted := 0
	for _, taskID := range validTasks {
		if err := taskService.DeleteTask(ctx, taskID); err != nil {
			printError("Failed to delete task %s: %v", taskID, err)
			continue
		}
		deleted++
	}

	if deleted > 0 {
		printSuccess("Successfully deleted %d task(s)", deleted)
	}

	if deleted < len(validTasks) {
		printWarning("Failed to delete %d task(s)", len(validTasks)-deleted)
	}

	return nil
}

func runDeleteAllTasks(ctx context.Context, force bool) error {
	if !force {
		printError("Deleting all tasks requires --force flag")
		return fmt.Errorf("--force flag required for bulk deletion")
	}

	// Get all tasks
	tasks, _, err := taskService.ListTasks(ctx, models.TaskFilter{}, "", "", 0, 10000)
	if err != nil {
		printError("Failed to get tasks: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks to delete")
		return nil
	}

	// Delete all tasks
	deleted := 0
	for _, task := range tasks {
		if err := taskService.DeleteTask(ctx, task.ID); err != nil {
			printError("Failed to delete task %s: %v", task.ID, err)
			continue
		}
		deleted++
	}

	printSuccess("Successfully deleted %d task(s)", deleted)
	return nil
}

func runDeleteTasksByStatus(ctx context.Context, status string, force bool) error {
	if !models.ValidateStatus(status) {
		printError("Invalid status: %s", status)
		return fmt.Errorf("invalid status")
	}

	// Get tasks with the specified status
	tasks, err := taskService.GetTasksByStatus(ctx, models.TaskStatus(status))
	if err != nil {
		printError("Failed to get tasks: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks found with status: %s", status)
		return nil
	}

	// Show confirmation
	if !force {
		fmt.Printf("The following %d task(s) with status '%s' will be deleted:\n", len(tasks), status)
		for _, task := range tasks {
			fmt.Printf("- %s: %s\n", task.ID[:8], task.Title)
		}
		fmt.Printf("\nAre you sure? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		
		if response != "y" && response != "yes" {
			printInfo("Deletion cancelled")
			return nil
		}
	}

	// Delete tasks
	deleted := 0
	for _, task := range tasks {
		if err := taskService.DeleteTask(ctx, task.ID); err != nil {
			printError("Failed to delete task %s: %v", task.ID, err)
			continue
		}
		deleted++
	}

	printSuccess("Successfully deleted %d task(s) with status '%s'", deleted, status)
	return nil
}

func runDeleteTasksByAssignee(ctx context.Context, assignee string, force bool) error {
	// Get tasks assigned to the specified user
	tasks, err := taskService.GetTasksByAssignee(ctx, assignee)
	if err != nil {
		printError("Failed to get tasks: %v", err)
		return err
	}

	if len(tasks) == 0 {
		printInfo("No tasks found assigned to: %s", assignee)
		return nil
	}

	// Show confirmation
	if !force {
		fmt.Printf("The following %d task(s) assigned to '%s' will be deleted:\n", len(tasks), assignee)
		for _, task := range tasks {
			fmt.Printf("- %s: %s\n", task.ID[:8], task.Title)
		}
		fmt.Printf("\nAre you sure? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		
		if response != "y" && response != "yes" {
			printInfo("Deletion cancelled")
			return nil
		}
	}

	// Delete tasks
	deleted := 0
	for _, task := range tasks {
		if err := taskService.DeleteTask(ctx, task.ID); err != nil {
			printError("Failed to delete task %s: %v", task.ID, err)
			continue
		}
		deleted++
	}

	printSuccess("Successfully deleted %d task(s) assigned to '%s'", deleted, assignee)
	return nil
}
