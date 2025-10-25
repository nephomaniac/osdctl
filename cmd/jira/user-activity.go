package jira

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const longUserActivityDescription = `
Query Jira for all tickets that contain a comment or update by a specific user within a given time period.

This command searches for Jira tickets where the specified user has:
- Added comments
- Updated the ticket
- Made any changes within the specified time period

You can specify the time period in two ways:
1. Using --days to search back N days from today
2. Using --start-date and --end-date to specify a custom date range

Requirements:
- JIRA_API_TOKEN environment variable or jira_token config setting

Examples:
  # Search last 7 days
  osdctl jira user-activity --user john.doe --days 7

  # Search with date range (format: YYYY-MM-DD)
  osdctl jira user-activity -u jane.smith --start-date 2025-01-01 --end-date 2025-01-31

  # Show detailed information
  osdctl jira user-activity --user john.doe --days 14 --detailed
`

type userActivityOptions struct {
	username   string
	days       int
	startDate  string
	endDate    string
	detailed   bool
	jiraToken  string
	jiraClient utils.JiraClientInterface

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
}

func newCmdUserActivity() *cobra.Command {
	ops := userActivityOptions{}

	var userActivityCmd = &cobra.Command{
		Use:   "user-activity",
		Short: "Query Jira tickets by user activity within a time period",
		Long:  longUserActivityDescription,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.QueryUserActivity())
		},
	}

	userActivityCmd.Flags().StringVarP(&ops.username, "user", "u", "", "Jira username to search for")
	userActivityCmd.Flags().IntVarP(&ops.days, "days", "d", 0, "Number of days to search back from today (mutually exclusive with start-date/end-date)")
	userActivityCmd.Flags().StringVar(&ops.startDate, "start-date", "", "Start date for search window (YYYY-MM-DD)")
	userActivityCmd.Flags().StringVar(&ops.endDate, "end-date", "", "End date for search window (YYYY-MM-DD)")
	userActivityCmd.Flags().BoolVarP(&ops.detailed, "detailed", "", false, "Show detailed ticket information including summaries")
	userActivityCmd.Flags().StringVarP(&ops.jiraToken, "jira-token", "t", "", "Override jira_token config and/or JIRA_API_TOKEN env var")

	_ = userActivityCmd.MarkFlagRequired("user")

	return userActivityCmd
}

// QueryUserActivity searches for Jira tickets with activity from the specified user
func (o *userActivityOptions) QueryUserActivity() error {
	var err error

	// Validate input
	if o.username == "" {
		return fmt.Errorf("username is required")
	}

	// Validate date options - either days OR date range, not both
	useDays := o.days > 0
	useDateRange := o.startDate != "" || o.endDate != ""

	if useDays && useDateRange {
		return fmt.Errorf("cannot use --days together with --start-date/--end-date; choose one approach")
	}

	if !useDays && !useDateRange {
		// Default to 30 days if nothing specified
		o.days = 30
		useDays = true
	}

	var startDate, endDate time.Time
	var jql string

	// Create JIRA client
	o.jiraClient, err = utils.NewJiraClient(o.jiraToken)
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Build JQL query based on mode
	if useDateRange {
		// Parse and validate date range
		startDate, endDate, err = o.validateDateRange()
		if err != nil {
			return err
		}

		jql = buildUserActivityJQLWithDateRange(o.username, startDate, endDate)
		fmt.Printf("Searching for tickets with activity by user '%s' from %s to %s...\n",
			o.username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	} else {
		// Use days-based query
		jql = buildUserActivityJQL(o.username, o.days)
		fmt.Printf("Searching for tickets with activity by user '%s' in the last %d days...\n", o.username, o.days)
	}

	fmt.Printf("JQL: %s\n\n", jql)

	// Search for issues
	issues, err := o.jiraClient.SearchIssues(jql)
	if err != nil {
		return fmt.Errorf("failed to search for issues: %w", err)
	}

	// Display results
	if len(issues) == 0 {
		if useDateRange {
			fmt.Printf("No tickets found with activity by user '%s' from %s to %s.\n",
				o.username, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		} else {
			fmt.Printf("No tickets found with activity by user '%s' in the last %d days.\n", o.username, o.days)
		}
		return nil
	}

	fmt.Printf("Found %d ticket(s) with activity by user '%s':\n\n", len(issues), o.username)

	if o.detailed {
		displayDetailedResults(issues)
	} else {
		displayBasicResults(issues)
	}

	return nil
}

// validateDateRange validates and parses the start and end dates
func (o *userActivityOptions) validateDateRange() (time.Time, time.Time, error) {
	const dateFormat = "2006-01-02"

	if o.startDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start-date is required when using date range")
	}
	if o.endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("end-date is required when using date range")
	}

	startDate, err := time.Parse(dateFormat, o.startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start-date format (expected YYYY-MM-DD): %w", err)
	}

	endDate, err := time.Parse(dateFormat, o.endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end-date format (expected YYYY-MM-DD): %w", err)
	}

	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("start-date cannot be after end-date")
	}

	return startDate, endDate, nil
}

// buildUserActivityJQL constructs a JQL query to find tickets with user activity
func buildUserActivityJQL(username string, days int) string {
	// JQL to find tickets where the user has activity within the time period
	// This searches for:
	// 1. Tickets with comments by the user (using regex match)
	// 2. Tickets that were updated in the given time period (filtered by the user's comments)
	// The tilde operator (~) performs a text search/contains match in Jira JQL
	// This allows partial username matching
	jql := fmt.Sprintf(
		`comment ~ "%s" AND updated >= -%dd ORDER BY updated DESC`,
		username,
		days,
	)
	return jql
}

// buildUserActivityJQLWithDateRange constructs a JQL query with specific date range
func buildUserActivityJQLWithDateRange(username string, startDate, endDate time.Time) string {
	// Jira JQL date format is YYYY-MM-DD or YYYY/MM/DD
	// We use the dash format for consistency
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Build JQL with date range
	// updated >= startDate AND updated <= endDate
	jql := fmt.Sprintf(
		`comment ~ "%s" AND updated >= "%s" AND updated <= "%s" ORDER BY updated DESC`,
		username,
		startDateStr,
		endDateStr,
	)
	return jql
}

// displayBasicResults shows a simple list of tickets
func displayBasicResults(issues []jira.Issue) {
	fmt.Printf("%-15s %-50s %-20s %-20s\n", "KEY", "SUMMARY", "STATUS", "LAST UPDATED")
	fmt.Println(strings.Repeat("-", 110))

	for _, issue := range issues {
		summary := issue.Fields.Summary
		if len(summary) > 47 {
			summary = summary[:47] + "..."
		}

		status := "Unknown"
		if issue.Fields.Status != nil {
			status = issue.Fields.Status.Name
		}

		updated := "Unknown"
		if !time.Time(issue.Fields.Updated).IsZero() {
			updated = time.Time(issue.Fields.Updated).Format("2006-01-02 15:04")
		}

		fmt.Printf("%-15s %-50s %-20s %-20s\n", issue.Key, summary, status, updated)
	}

	fmt.Printf("\nTotal: %d ticket(s)\n", len(issues))
}

// displayDetailedResults shows comprehensive information about each ticket
func displayDetailedResults(issues []jira.Issue) {
	for i, issue := range issues {
		fmt.Printf("[%d] %s\n", i+1, strings.Repeat("=", 80))
		fmt.Printf("Key:     %s\n", issue.Key)
		fmt.Printf("Summary: %s\n", issue.Fields.Summary)

		if issue.Fields.Status != nil {
			fmt.Printf("Status:  %s\n", issue.Fields.Status.Name)
		}

		if issue.Fields.Assignee != nil {
			fmt.Printf("Assignee: %s\n", issue.Fields.Assignee.DisplayName)
		}

		if !time.Time(issue.Fields.Updated).IsZero() {
			fmt.Printf("Updated: %s\n", time.Time(issue.Fields.Updated).Format("2006-01-02 15:04:05"))
		}

		if !time.Time(issue.Fields.Created).IsZero() {
			fmt.Printf("Created: %s\n", time.Time(issue.Fields.Created).Format("2006-01-02 15:04:05"))
		}

		// Show URL
		jiraBaseURL := os.Getenv("JIRA_BASE_URL")
		if jiraBaseURL == "" {
			jiraBaseURL = utils.JiraBaseURL
		}
		fmt.Printf("URL:     %s/browse/%s\n", jiraBaseURL, issue.Key)

		// Show description preview
		if issue.Fields.Description != "" {
			desc := issue.Fields.Description
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}
			fmt.Printf("\nDescription:\n%s\n", desc)
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d ticket(s)\n", len(issues))
}
