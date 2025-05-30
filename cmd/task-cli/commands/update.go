package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [task-id]",
	Short: "Update an existing task",
	Long: `Update an existing task with new values for any field.

You can update the title, description, status, priority, assignee, due date, and tags.
Only the fields you specify will be updated; others will remain unchanged.

Examples:
  task-cli update 123 --title "New title"
  task-cli update 123 --status completed
  task-cli update 123 --priority high --assignee john
  task-cli update 123 --due "2024-01-20 15:00"
  task-cli update 123 --tags "bug,urgent,frontend"
  task-cli update 123 --description "Updated description"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdateTask(cmd, args)
	},
}

func init() {
	updateCmd.Flags().String("title", "", "Update task title")
	updateCmd.Flags().StringP("description", "d", "", "Update task description")
	updateCmd.Flags().StringP("status", "s", "", "Update task status (pending,in-progress,completed,cancelled,on-hold)")
	updateCmd.Flags().StringP("priority", "p", "", "Update task priority (low,medium,high,critical)")
	updateCmd.Flags().StringP("assignee", "a", "", "Update task assignee")
	updateCmd.Flags().String("due", "", "Update due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "Update task tags (comma-separated)")
	updateCmd.Flags().Bool("clear-due", false, "Clear the due date")
	updateCmd.Flags().Bool("clear-assignee", false, "Clear the assignee")
	updateCmd.Flags().Bool("clear-tags", false, "Clear all tags")
	updateCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for task update")
}

func runUpdateTask(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	taskID := args[0]

	// Get current task
	currentTask, err := taskService.GetTask(ctx, taskID)
	if err != nil {
		printError("Failed to get task: %v", err)
		return err
	}

	// Build update request
	req, err := buildUpdateRequest(cmd, currentTask)
	if err != nil {
		printError("Failed to build update request: %v", err)
		return err
	}

	// Check if any updates were provided
	if !hasUpdates(req) {
		printWarning("No updates provided")
		return nil
	}

	// Update task
	updatedTask, err := taskService.UpdateTask(ctx, taskID, req)
	if err != nil {
		printError("Failed to update task: %v", err)
		return err
	}

	// Print success message
	printSuccess("Task updated successfully!")
	printTaskDetails(updatedTask)

	return nil
}

func buildUpdateRequest(cmd *cobra.Command, currentTask *models.Task) (models.TaskUpdateRequest, error) {
	var req models.TaskUpdateRequest

	interactive, _ := cmd.Flags().GetBool("interactive")

	// Update title
	if title, _ := cmd.Flags().GetString("title"); title != "" {
		req.Title = &title
	} else if interactive {
		if newTitle := promptForUpdate("Title", currentTask.Title); newTitle != currentTask.Title {
			req.Title = &newTitle
		}
	}

	// Update description
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		req.Description = &description
	} else if interactive {
		if newDesc := promptForUpdate("Description", currentTask.Description); newDesc != currentTask.Description {
			req.Description = &newDesc
		}
	}

	// Update status
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		if !models.ValidateStatus(status) {
			return req, fmt.Errorf("invalid status: %s", status)
		}
		taskStatus := models.TaskStatus(status)
		req.Status = &taskStatus
	} else if interactive {
		if newStatus := promptForStatusUpdate(currentTask.Status); newStatus != currentTask.Status {
			req.Status = &newStatus
		}
	}

	// Update priority
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		if !models.ValidatePriority(priority) {
			return req, fmt.Errorf("invalid priority: %s", priority)
		}
		taskPriority := models.TaskPriority(priority)
		req.Priority = &taskPriority
	} else if interactive {
		if newPriority := promptForPriorityUpdate(currentTask.Priority); newPriority != currentTask.Priority {
			req.Priority = &newPriority
		}
	}

	// Update assignee
	if clearAssignee, _ := cmd.Flags().GetBool("clear-assignee"); clearAssignee {
		empty := ""
		req.Assignee = &empty
	} else if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		req.Assignee = &assignee
	} else if interactive {
		if newAssignee := promptForUpdate("Assignee", currentTask.Assignee); newAssignee != currentTask.Assignee {
			req.Assignee = &newAssignee
		}
	}

	// Update due date
	if clearDue, _ := cmd.Flags().GetBool("clear-due"); clearDue {
		req.DueDate = nil
	} else if dueStr, _ := cmd.Flags().GetString("due"); dueStr != "" {
		dueDate, err := parseDueDate(dueStr)
		if err != nil {
			return req, fmt.Errorf("invalid due date format: %v", err)
		}
		req.DueDate = dueDate
	} else if interactive {
		currentDueStr := ""
		if currentTask.DueDate != nil {
			currentDueStr = currentTask.DueDate.Format("2006-01-02 15:04")
		}
		if newDueStr := promptForUpdate("Due date (YYYY-MM-DD HH:MM)", currentDueStr); newDueStr != currentDueStr {
			if newDueStr == "" {
				req.DueDate = nil
			} else {
				dueDate, err := parseDueDate(newDueStr)
				if err != nil {
					return req, fmt.Errorf("invalid due date format: %v", err)
				}
				req.DueDate = dueDate
			}
		}
	}

	// Update tags
	if clearTags, _ := cmd.Flags().GetBool("clear-tags"); clearTags {
		req.Tags = []string{}
	} else if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		req.Tags = tags
	} else if interactive {
		currentTagsStr := strings.Join(currentTask.Tags, ", ")
		if newTagsStr := promptForUpdate("Tags (comma-separated)", currentTagsStr); newTagsStr != currentTagsStr {
			if newTagsStr == "" {
				req.Tags = []string{}
			} else {
				req.Tags = strings.Split(newTagsStr, ",")
				for i, tag := range req.Tags {
					req.Tags[i] = strings.TrimSpace(tag)
				}
			}
		}
	}

	return req, nil
}

func hasUpdates(req models.TaskUpdateRequest) bool {
	return req.Title != nil ||
		req.Description != nil ||
		req.Status != nil ||
		req.Priority != nil ||
		req.Assignee != nil ||
		req.DueDate != nil ||
		req.Tags != nil
}

func promptForUpdate(field, currentValue string) string {
	fmt.Printf("%s [%s]: ", field, currentValue)
	var input string
	fmt.Scanln(&input)
	
	input = strings.TrimSpace(input)
	if input == "" {
		return currentValue
	}
	return input
}

func promptForStatusUpdate(currentStatus models.TaskStatus) models.TaskStatus {
	statuses := models.GetAllStatuses()
	
	fmt.Printf("Current status: %s\n", currentStatus)
	fmt.Println("Available statuses:")
	for i, status := range statuses {
		marker := " "
		if status == currentStatus {
			marker = "*"
		}
		fmt.Printf(" %s %d. %s\n", marker, i+1, status)
	}

	fmt.Printf("Select new status (1-%d) [current]: ", len(statuses))
	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)
	if input == "" {
		return currentStatus
	}

	// Try to parse as number
	switch input {
	case "1":
		return models.StatusPending
	case "2":
		return models.StatusInProgress
	case "3":
		return models.StatusCompleted
	case "4":
		return models.StatusCancelled
	case "5":
		return models.StatusOnHold
	}

	// Try to parse as string
	if models.ValidateStatus(input) {
		return models.TaskStatus(input)
	}

	return currentStatus
}

func promptForPriorityUpdate(currentPriority models.TaskPriority) models.TaskPriority {
	priorities := models.GetAllPriorities()
	
	fmt.Printf("Current priority: %s\n", currentPriority)
	fmt.Println("Available priorities:")
	for i, priority := range priorities {
		marker := " "
		if priority == currentPriority {
			marker = "*"
		}
		fmt.Printf(" %s %d. %s\n", marker, i+1, priority)
	}

	fmt.Printf("Select new priority (1-%d) [current]: ", len(priorities))
	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)
	if input == "" {
		return currentPriority
	}

	// Try to parse as number
	switch input {
	case "1":
		return models.PriorityLow
	case "2":
		return models.PriorityMedium
	case "3":
		return models.PriorityHigh
	case "4":
		return models.PriorityCritical
	}

	// Try to parse as string
	if models.ValidatePriority(input) {
		return models.TaskPriority(input)
	}

	return currentPriority
}
