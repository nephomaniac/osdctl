package github

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v63/github"
	"github.com/manifoldco/promptui"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const longReviewDescription = `
Review a GitHub Pull Request using AI to generate a comprehensive code review.

This command fetches the PR details, analyzes the changes using AI, and generates a structured review including:
- Summary of changes
- Code quality assessment
- Potential issues and suggestions
- Security considerations
- Testing recommendations

After generating the review, you'll be prompted to optionally post it as a comment on the PR.

Requirements:
- GITHUB_TOKEN environment variable (GitHub Personal Access Token with repo access)
- OPENAI_API_KEY environment variable (API key for AI service)
- MODEL_PROVIDER_BASE_URL environment variable (optional, defaults to http://localhost:11434/v1)
- MODEL_NAME environment variable (optional, defaults to mistral-small)

Examples:
  osdctl github review-pr https://github.com/openshift/osdctl/pull/794
  osdctl github review-pr https://github.com/owner/repo/pull/123 --auto-post
  osdctl github review-pr https://github.com/owner/repo/pull/123 --model gpt-4
`

var reviewPRCmd = &cobra.Command{
	Use:   "review-pr <github-pr-url>",
	Short: "Generate an AI-powered review of a GitHub Pull Request",
	Long:  longReviewDescription,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one argument: GitHub PR URL")
		}
		// Validate that it looks like a GitHub PR URL
		if !strings.Contains(args[0], "github.com") || !strings.Contains(args[0], "/pull/") {
			return fmt.Errorf("invalid GitHub PR URL. Expected format: https://github.com/owner/repo/pull/number")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		prURL := args[0]

		autoPost, _ := cmd.Flags().GetBool("auto-post")
		modelName, _ := cmd.Flags().GetString("model")
		baseURL, _ := cmd.Flags().GetString("base-url")
		skipPost, _ := cmd.Flags().GetBool("skip-post")

		return ReviewPullRequest(prURL, autoPost, skipPost, modelName, baseURL)
	},
}

func init() {
	reviewPRCmd.Flags().Bool("auto-post", false, "Automatically post the review as a comment without prompting")
	reviewPRCmd.Flags().Bool("skip-post", false, "Skip posting the review (just display it)")
	reviewPRCmd.Flags().String("model", "", "AI model name (overrides MODEL_NAME env var)")
	reviewPRCmd.Flags().String("base-url", "", "AI model provider base URL (overrides MODEL_PROVIDER_BASE_URL env var)")
}

// PRInfo holds parsed pull request information
type PRInfo struct {
	Owner  string
	Repo   string
	Number int
}

// parsePRURL extracts owner, repo, and PR number from a GitHub PR URL
func parsePRURL(url string) (*PRInfo, error) {
	// Support multiple URL formats:
	// https://github.com/owner/repo/pull/123
	// github.com/owner/repo/pull/123
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid GitHub PR URL format. Expected: https://github.com/owner/repo/pull/number")
	}

	var number int
	_, err := fmt.Sscanf(matches[3], "%d", &number)
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %w", err)
	}

	return &PRInfo{
		Owner:  matches[1],
		Repo:   matches[2],
		Number: number,
	}, nil
}

// ReviewPullRequest fetches and reviews a GitHub pull request
func ReviewPullRequest(prURL string, autoPost, skipPost bool, modelName, baseURL string) error {
	// Validate environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Set AI model defaults
	if modelName == "" {
		modelName = os.Getenv("MODEL_NAME")
		if modelName == "" {
			modelName = "mistral-small"
		}
	}

	if baseURL == "" {
		baseURL = os.Getenv("MODEL_PROVIDER_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
	}

	// Parse PR URL
	prInfo, err := parsePRURL(prURL)
	if err != nil {
		return err
	}

	fmt.Printf("Fetching PR #%d from %s/%s...\n", prInfo.Number, prInfo.Owner, prInfo.Repo)

	// Create GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Fetch PR details
	pr, _, err := client.PullRequests.Get(ctx, prInfo.Owner, prInfo.Repo, prInfo.Number)
	if err != nil {
		return fmt.Errorf("failed to fetch PR: %w", err)
	}

	// Fetch PR files/changes
	fmt.Printf("Fetching changes for PR #%d...\n", prInfo.Number)
	files, _, err := client.PullRequests.ListFiles(ctx, prInfo.Owner, prInfo.Repo, prInfo.Number, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch PR files: %w", err)
	}

	// Fetch PR diff
	diff, _, err := client.PullRequests.GetRaw(ctx, prInfo.Owner, prInfo.Repo, prInfo.Number, github.RawOptions{Type: github.Diff})
	if err != nil {
		return fmt.Errorf("failed to fetch PR diff: %w", err)
	}

	fmt.Printf("Analyzing %d file(s) with AI model %s...\n", len(files), modelName)

	// Generate AI review
	reviewer := NewPRReviewer(apiKey, modelName, baseURL)
	review, err := reviewer.ReviewPR(pr, files, diff)
	if err != nil {
		return fmt.Errorf("failed to generate review: %w", err)
	}

	// Display the review
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("AI-Generated Review for PR #%d: %s\n", prInfo.Number, pr.GetTitle())
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(review)
	fmt.Println(strings.Repeat("=", 80))

	// Handle posting the review
	if skipPost {
		fmt.Println("\nSkipping post (--skip-post flag set)")
		return nil
	}

	shouldPost := autoPost
	if !autoPost {
		// Prompt user
		prompt := promptui.Prompt{
			Label:     "Would you like to post this review as a comment to the PR? (yes/no)",
			IsConfirm: true,
		}

		result, err := prompt.Run()
		if err != nil && err != promptui.ErrAbort {
			return fmt.Errorf("prompt failed: %w", err)
		}

		shouldPost = (err == nil && strings.ToLower(result) == "yes")
	}

	if shouldPost {
		fmt.Printf("\nPosting review to PR #%d...\n", prInfo.Number)
		err = postReviewComment(ctx, client, prInfo, review)
		if err != nil {
			return fmt.Errorf("failed to post review: %w", err)
		}
		fmt.Printf("✓ Review successfully posted to %s\n", prURL)
	} else {
		fmt.Println("\nReview not posted.")
	}

	return nil
}

// postReviewComment posts the review as a comment on the PR
func postReviewComment(ctx context.Context, client *github.Client, prInfo *PRInfo, review string) error {
	comment := &github.IssueComment{
		Body: github.String(review),
	}

	_, _, err := client.Issues.CreateComment(ctx, prInfo.Owner, prInfo.Repo, prInfo.Number, comment)
	return err
}

// PRReviewer handles AI-powered PR reviews
type PRReviewer struct {
	aiClient *utils.OpenAIClient
	model    string
}

// NewPRReviewer creates a new PR reviewer
func NewPRReviewer(apiKey, modelName, baseURL string) *PRReviewer {
	return &PRReviewer{
		aiClient: utils.NewOpenAIClient(baseURL, apiKey),
		model:    modelName,
	}
}

// ReviewPR generates an AI review of a pull request
func (r *PRReviewer) ReviewPR(pr *github.PullRequest, files []*github.CommitFile, diff string) (string, error) {
	// Build context for AI
	context := r.buildPRContext(pr, files, diff)

	// Call AI with structured review template
	systemPrompt := `You are an expert code reviewer with deep knowledge of software engineering best practices, security, testing, and maintainability. Your task is to provide comprehensive, constructive code reviews that help improve code quality and catch potential issues.

Focus on:
- Code quality and maintainability
- Potential bugs or logic errors
- Security vulnerabilities
- Performance concerns
- Testing coverage and recommendations
- Architectural considerations
- Best practices adherence

Be thorough but constructive. Provide specific, actionable feedback.`

	userPrompt := fmt.Sprintf(`Please review the following GitHub Pull Request and provide a comprehensive code review.

%s

Please provide your review in this structured format:

## Overview
[Brief summary of what this PR does and your overall assessment]

## Strengths
[List positive aspects of the changes]

## Potential Issues & Suggestions
[Detailed list of issues, concerns, or improvements organized by category]

### Code Quality
[Issues related to code structure, readability, maintainability]

### Security Considerations
[Any security concerns or vulnerabilities]

### Testing
[Testing recommendations or concerns]

### Performance
[Performance-related observations]

### Documentation
[Documentation needs or improvements]

## Recommendation
[Your overall recommendation: APPROVE, REQUEST CHANGES, or COMMENT with reasoning]

Be specific and provide examples where helpful. Keep the tone professional and constructive.`, context)

	review, err := r.aiClient.ChatCompletion(systemPrompt, userPrompt, r.model)
	if err != nil {
		return "", err
	}

	return review, nil
}

// buildPRContext builds a comprehensive context string for AI analysis
func (r *PRReviewer) buildPRContext(pr *github.PullRequest, files []*github.CommitFile, diff string) string {
	var sb strings.Builder

	// PR metadata
	sb.WriteString(fmt.Sprintf("**Title**: %s\n", pr.GetTitle()))
	sb.WriteString(fmt.Sprintf("**Author**: %s\n", pr.GetUser().GetLogin()))
	sb.WriteString(fmt.Sprintf("**State**: %s\n", pr.GetState()))
	sb.WriteString(fmt.Sprintf("**Additions**: +%d lines\n", pr.GetAdditions()))
	sb.WriteString(fmt.Sprintf("**Deletions**: -%d lines\n", pr.GetDeletions()))
	sb.WriteString(fmt.Sprintf("**Files Changed**: %d\n\n", len(files)))

	// PR description
	if pr.Body != nil && *pr.Body != "" {
		sb.WriteString(fmt.Sprintf("**Description**:\n%s\n\n", *pr.Body))
	}

	// File changes summary
	sb.WriteString("**Files Changed**:\n")
	for _, file := range files {
		status := file.GetStatus()
		sb.WriteString(fmt.Sprintf("- %s (%s): +%d -%d\n",
			file.GetFilename(),
			status,
			file.GetAdditions(),
			file.GetDeletions(),
		))
	}
	sb.WriteString("\n")

	// Include the diff (truncate if too large)
	maxDiffSize := 50000 // ~50KB limit
	sb.WriteString("**Code Changes (Diff)**:\n```diff\n")
	if len(diff) > maxDiffSize {
		sb.WriteString(diff[:maxDiffSize])
		sb.WriteString("\n... [diff truncated for length] ...\n")
	} else {
		sb.WriteString(diff)
	}
	sb.WriteString("\n```\n")

	return sb.String()
}
