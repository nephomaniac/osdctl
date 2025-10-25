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
- Made any changes within the specified number of days

Requirements:
- JIRA_API_TOKEN environment variable or jira_token config setting

Example:
  osdctl jira user-activity --user john.doe --days 7
  osdctl jira user-activity -u jane.smith -d 30
  osdctl jira user-activity --user john.doe --days 14 --detailed
`

type userActivityOptions struct {
	username   string
	days       int
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
	userActivityCmd.Flags().IntVarP(&ops.days, "days", "d", 30, "Number of days to search back from today")
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

	if o.days <= 0 {
		return fmt.Errorf("days must be a positive number")
	}

	// Create JIRA client
	o.jiraClient, err = utils.NewJiraClient(o.jiraToken)
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Build JQL query
	jql := buildUserActivityJQL(o.username, o.days)
	fmt.Printf("Searching for tickets with activity by user '%s' in the last %d days...\n", o.username, o.days)
	fmt.Printf("JQL: %s\n\n", jql)

	// Search for issues
	issues, err := o.jiraClient.SearchIssues(jql)
	if err != nil {
		return fmt.Errorf("failed to search for issues: %w", err)
	}

	// Display results
	if len(issues) == 0 {
		fmt.Printf("No tickets found with activity by user '%s' in the last %d days.\n", o.username, o.days)
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
