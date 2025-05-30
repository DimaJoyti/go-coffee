package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display task statistics and analytics",
	Long: `Display comprehensive statistics about your tasks including:
  • Total task count
  • Breakdown by status (pending, in-progress, completed, etc.)
  • Breakdown by priority (low, medium, high, critical)
  • Breakdown by assignee
  • Overdue tasks count
  • Tasks due today and this week

Examples:
  task-cli stats
  task-cli stats --output json
  task-cli stats --detailed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTaskStats(cmd, args)
	},
}

func init() {
	statsCmd.Flags().Bool("detailed", false, "Show detailed statistics")
	statsCmd.Flags().Bool("chart", false, "Display ASCII charts")
}

func runTaskStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	detailed, _ := cmd.Flags().GetBool("detailed")
	showChart, _ := cmd.Flags().GetBool("chart")

	// Get statistics
	stats, err := taskService.GetTaskStats(ctx)
	if err != nil {
		printError("Failed to get task statistics: %v", err)
		return err
	}

	// Display based on output format
	outputFormat := getOutputFormat()
	switch outputFormat {
	case "json":
		return printStatsJSON(stats)
	case "yaml":
		return printStatsYAML(stats)
	default:
		printStatsTable(stats, detailed, showChart)
	}

	return nil
}

func printStatsJSON(stats *models.TaskStats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats to JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printStatsYAML(stats *models.TaskStats) error {
	data, err := yaml.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats to YAML: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printStatsTable(stats *models.TaskStats, detailed, showChart bool) {
	if isColorEnabled() {
		fmt.Printf("\n📊 \033[1mTask Statistics\033[0m\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	} else {
		fmt.Printf("\nTask Statistics\n")
		fmt.Printf("===============\n\n")
	}

	// Overall statistics
	printOverallStats(stats)

	// Status breakdown
	printStatusBreakdown(stats, showChart)

	// Priority breakdown
	printPriorityBreakdown(stats, showChart)

	// Due date statistics
	printDueDateStats(stats)

	// Assignee breakdown (if detailed)
	if detailed {
		printAssigneeBreakdown(stats, showChart)
	}

	fmt.Println()
}

func printOverallStats(stats *models.TaskStats) {
	if isColorEnabled() {
		fmt.Printf("📋 \033[1mOverall\033[0m\n")
		fmt.Printf("  Total Tasks: \033[36m%d\033[0m\n", stats.Total)
		fmt.Printf("  Overdue: \033[31m%d\033[0m\n", stats.Overdue)
		fmt.Printf("  Due Today: \033[33m%d\033[0m\n", stats.DueToday)
		fmt.Printf("  Due This Week: \033[34m%d\033[0m\n\n", stats.DueThisWeek)
	} else {
		fmt.Printf("Overall:\n")
		fmt.Printf("  Total Tasks: %d\n", stats.Total)
		fmt.Printf("  Overdue: %d\n", stats.Overdue)
		fmt.Printf("  Due Today: %d\n", stats.DueToday)
		fmt.Printf("  Due This Week: %d\n\n", stats.DueThisWeek)
	}
}

func printStatusBreakdown(stats *models.TaskStats, showChart bool) {
	if isColorEnabled() {
		fmt.Printf("📈 \033[1mBy Status\033[0m\n")
	} else {
		fmt.Printf("By Status:\n")
	}

	// Sort statuses by count
	type statusCount struct {
		status models.TaskStatus
		count  int
	}
	var statusCounts []statusCount
	for status, count := range stats.ByStatus {
		statusCounts = append(statusCounts, statusCount{status, count})
	}
	sort.Slice(statusCounts, func(i, j int) bool {
		return statusCounts[i].count > statusCounts[j].count
	})

	for _, sc := range statusCounts {
		percentage := float64(sc.count) / float64(stats.Total) * 100
		if isColorEnabled() {
			statusColor := getStatusColor(sc.status)
			fmt.Printf("  %s%-12s\033[0m: \033[36m%3d\033[0m (\033[35m%5.1f%%\033[0m)",
				statusColor, sc.status, sc.count, percentage)
		} else {
			fmt.Printf("  %-12s: %3d (%5.1f%%)", sc.status, sc.count, percentage)
		}

		if showChart && stats.Total > 0 {
			barLength := int(percentage / 2) // Scale to max 50 chars
			if barLength > 0 {
				fmt.Printf(" %s", strings.Repeat("█", barLength))
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func printPriorityBreakdown(stats *models.TaskStats, showChart bool) {
	if isColorEnabled() {
		fmt.Printf("🎯 \033[1mBy Priority\033[0m\n")
	} else {
		fmt.Printf("By Priority:\n")
	}

	// Sort priorities by importance
	priorities := []models.TaskPriority{
		models.PriorityCritical,
		models.PriorityHigh,
		models.PriorityMedium,
		models.PriorityLow,
	}

	for _, priority := range priorities {
		count := stats.ByPriority[priority]
		percentage := float64(count) / float64(stats.Total) * 100
		if isColorEnabled() {
			priorityColor := getPriorityColor(priority)
			fmt.Printf("  %s%-12s\033[0m: \033[36m%3d\033[0m (\033[35m%5.1f%%\033[0m)",
				priorityColor, priority, count, percentage)
		} else {
			fmt.Printf("  %-12s: %3d (%5.1f%%)", priority, count, percentage)
		}

		if showChart && stats.Total > 0 {
			barLength := int(percentage / 2) // Scale to max 50 chars
			if barLength > 0 {
				fmt.Printf(" %s", strings.Repeat("█", barLength))
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func printDueDateStats(stats *models.TaskStats) {
	if isColorEnabled() {
		fmt.Printf("📅 \033[1mDue Dates\033[0m\n")
		fmt.Printf("  \033[31mOverdue\033[0m: \033[36m%d\033[0m\n", stats.Overdue)
		fmt.Printf("  \033[33mDue Today\033[0m: \033[36m%d\033[0m\n", stats.DueToday)
		fmt.Printf("  \033[34mDue This Week\033[0m: \033[36m%d\033[0m\n\n", stats.DueThisWeek)
	} else {
		fmt.Printf("Due Dates:\n")
		fmt.Printf("  Overdue: %d\n", stats.Overdue)
		fmt.Printf("  Due Today: %d\n", stats.DueToday)
		fmt.Printf("  Due This Week: %d\n\n", stats.DueThisWeek)
	}
}

func printAssigneeBreakdown(stats *models.TaskStats, showChart bool) {
	if len(stats.ByAssignee) == 0 {
		return
	}

	if isColorEnabled() {
		fmt.Printf("👥 \033[1mBy Assignee\033[0m\n")
	} else {
		fmt.Printf("By Assignee:\n")
	}

	// Sort assignees by count
	type assigneeCount struct {
		assignee string
		count    int
	}
	var assigneeCounts []assigneeCount
	for assignee, count := range stats.ByAssignee {
		assigneeCounts = append(assigneeCounts, assigneeCount{assignee, count})
	}
	sort.Slice(assigneeCounts, func(i, j int) bool {
		return assigneeCounts[i].count > assigneeCounts[j].count
	})

	// Show top 10 assignees
	maxShow := 10
	if len(assigneeCounts) < maxShow {
		maxShow = len(assigneeCounts)
	}

	for i := 0; i < maxShow; i++ {
		ac := assigneeCounts[i]
		percentage := float64(ac.count) / float64(stats.Total) * 100
		
		assignee := ac.assignee
		if len(assignee) > 20 {
			assignee = assignee[:17] + "..."
		}

		if isColorEnabled() {
			fmt.Printf("  \033[32m%-20s\033[0m: \033[36m%3d\033[0m (\033[35m%5.1f%%\033[0m)",
				assignee, ac.count, percentage)
		} else {
			fmt.Printf("  %-20s: %3d (%5.1f%%)", assignee, ac.count, percentage)
		}

		if showChart && stats.Total > 0 {
			barLength := int(percentage / 2) // Scale to max 50 chars
			if barLength > 0 {
				fmt.Printf(" %s", strings.Repeat("█", barLength))
			}
		}
		fmt.Println()
	}

	if len(assigneeCounts) > maxShow {
		remaining := len(assigneeCounts) - maxShow
		if isColorEnabled() {
			fmt.Printf("  \033[90m... and %d more\033[0m\n", remaining)
		} else {
			fmt.Printf("  ... and %d more\n", remaining)
		}
	}
	fmt.Println()
}
