package jira

import (
	"fmt"
	"os"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
)

const longSummarizeDescription = `
Summarize a JIRA ticket using AI to generate a structured summary of the ticket's comments and activity.

This command analyzes the ticket's description and all comments to provide a comprehensive summary including:
- How long the ticket has been open
- Active research days by SREs
- Actions taken and knowledge gained
- Current hypothesis and next steps

The summary is formatted using JIRA wiki markup and can be posted as a comment to the ticket.

Requirements:
- OPENAI_API_KEY environment variable (API key for AI service)
- MODEL_PROVIDER_BASE_URL environment variable (optional, defaults to http://localhost:11434/v1)
- MODEL_NAME environment variable (optional, defaults to mistral-small)

Example:
  osdctl jira summarize SREP-12345
  osdctl jira summarize SREP-12345 --post-comment
  osdctl jira summarize SREP-12345 --model gpt-4 --base-url https://api.openai.com/v1
`

var summarizeCmd = &cobra.Command{
	Use:   "summarize <ticket-key>",
	Short: "Generate an AI-powered summary of a JIRA ticket",
	Long:  longSummarizeDescription,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticketKey := strings.ToUpper(args[0])

		postComment, _ := cmd.Flags().GetBool("post-comment")
		modelName, _ := cmd.Flags().GetString("model")
		baseURL, _ := cmd.Flags().GetString("base-url")
		commentThreshold, _ := cmd.Flags().GetInt("comment-threshold")

		return SummarizeTicket(ticketKey, postComment, modelName, baseURL, commentThreshold)
	},
}

func init() {
	summarizeCmd.Flags().Bool("post-comment", false, "Post the summary as a comment to the JIRA ticket")
	summarizeCmd.Flags().String("model", "", "AI model name (overrides MODEL_NAME env var)")
	summarizeCmd.Flags().String("base-url", "", "AI model provider base URL (overrides MODEL_PROVIDER_BASE_URL env var)")
	summarizeCmd.Flags().Int("comment-threshold", 5, "Minimum number of comments required to generate summary")
}

// SummarizeTicket generates an AI-powered summary of a JIRA ticket
func SummarizeTicket(ticketKey string, postComment bool, modelName, baseURL string, commentThreshold int) error {
	// Validate environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Set defaults from environment or flags
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

	// Create JIRA client
	jiraClient, err := utils.NewJiraClient("")
	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Fetch the ticket
	fmt.Printf("Fetching ticket %s...\n", ticketKey)
	issue, err := getIssue(jiraClient, ticketKey)
	if err != nil {
		return fmt.Errorf("failed to fetch ticket: %w", err)
	}

	// Get all comments
	comments, err := getComments(jiraClient, ticketKey)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	fmt.Printf("Found %d comments on ticket %s\n", len(comments), ticketKey)

	// Check comment threshold
	if len(comments) < commentThreshold {
		return fmt.Errorf("ticket has only %d comments (threshold is %d)", len(comments), commentThreshold)
	}

	// Initialize AI summarizer
	summarizer := NewCommentSummarizer(apiKey, modelName, baseURL, commentThreshold)

	// Generate summary
	fmt.Printf("Generating AI summary using model %s...\n", modelName)
	summary, err := summarizer.SummarizeComments(comments, ticketKey, issue.Fields.Description)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	if summary == "" {
		return fmt.Errorf("no summary generated")
	}

	// Format the final output
	finalSummary := formatSummaryForJira(summary, ticketKey, len(comments))

	// Print summary to console
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("AI-Generated Summary for", ticketKey)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(finalSummary)
	fmt.Println(strings.Repeat("=", 80))

	// Optionally post as comment
	if postComment {
		fmt.Printf("\nPosting summary as comment to %s...\n", ticketKey)
		err = postCommentToJira(jiraClient, ticketKey, finalSummary)
		if err != nil {
			return fmt.Errorf("failed to post comment: %w", err)
		}
		fmt.Printf("Summary successfully posted to %s/browse/%s\n", utils.JiraBaseURL, ticketKey)
	}

	return nil
}

// getIssue retrieves a JIRA issue by key
func getIssue(client utils.JiraClientInterface, issueKey string) (*jira.Issue, error) {
	issue, _, err := client.Issue().Get(issueKey, nil)
	return issue, err
}

// getComments retrieves all comments for a JIRA issue
func getComments(client utils.JiraClientInterface, issueKey string) ([]*jira.Comment, error) {
	issue, _, err := client.Issue().Get(issueKey, nil)
	if err != nil {
		return nil, err
	}

	if issue.Fields == nil || issue.Fields.Comments == nil {
		return []*jira.Comment{}, nil
	}

	return issue.Fields.Comments.Comments, nil
}

// postCommentToJira posts a comment to a JIRA issue
func postCommentToJira(client utils.JiraClientInterface, issueKey, commentBody string) error {
	comment := &jira.Comment{
		Body: commentBody,
	}

	_, _, err := client.Issue().AddComment(issueKey, comment)
	return err
}

// formatSummaryForJira formats the AI-generated summary with JIRA-specific formatting
func formatSummaryForJira(summary, ticketKey string, commentCount int) string {
	header := fmt.Sprintf("h2. AI-Generated Ticket Summary\n\n")
	header += fmt.Sprintf("This summary was automatically generated for ticket %s based on %d comments.\n\n", ticketKey, commentCount)
	header += "----\n\n"

	footer := "\n\n----\n"
	footer += "_This summary was automatically generated using AI analysis of the ticket's comments and description._"

	return header + summary + footer
}

// CommentSummarizer handles AI-powered comment summarization
type CommentSummarizer struct {
	apiKey           string
	modelName        string
	baseURL          string
	commentThreshold int
}

// NewCommentSummarizer creates a new CommentSummarizer
func NewCommentSummarizer(apiKey, modelName, baseURL string, commentThreshold int) *CommentSummarizer {
	return &CommentSummarizer{
		apiKey:           apiKey,
		modelName:        modelName,
		baseURL:          baseURL,
		commentThreshold: commentThreshold,
	}
}

// SummarizeComments generates an AI summary of JIRA comments
func (s *CommentSummarizer) SummarizeComments(comments []*jira.Comment, ticketKey, description string) (string, error) {
	if len(comments) < s.commentThreshold {
		return "", fmt.Errorf("insufficient comments (%d < %d)", len(comments), s.commentThreshold)
	}

	// Format comments for AI processing
	formattedComments := s.formatCommentsForAI(comments, description)

	// Generate summary using AI
	summary, err := s.callAI(formattedComments, ticketKey)
	if err != nil {
		return "", err
	}

	return summary, nil
}

// formatCommentsForAI formats JIRA comments for AI processing
func (s *CommentSummarizer) formatCommentsForAI(comments []*jira.Comment, description string) string {
	var parts []string

	// Include original ticket description
	if description != "" {
		parts = append(parts, fmt.Sprintf("ORIGINAL TICKET DESCRIPTION:\n%s\n", description))
		parts = append(parts, strings.Repeat("=", 50))
	}

	// Add all comments
	for i, comment := range comments {
		author := "Unknown"
		if comment.Author.DisplayName != "" {
			author = comment.Author.DisplayName
		}

		created := comment.Created
		if len(created) > 10 {
			created = created[:10] // Just the date part
		}

		parts = append(parts, fmt.Sprintf("Comment %d (%s, %s):\n%s\n", i+1, author, created, comment.Body))
	}

	return strings.Join(parts, "\n")
}

// callAI makes the API call to the AI service
func (s *CommentSummarizer) callAI(formattedComments, ticketKey string) (string, error) {
	// Import OpenAI library
	client := utils.NewOpenAIClient(s.baseURL, s.apiKey)

	systemPrompt := `You are an expert SRE analyst reviewing JIRA support tickets. Your task is to create concise summaries that help SREs quickly understand ticket status and next steps. Focus on technical details, research efforts, hypotheses, and actionable next steps. Be brief but comprehensive - these summaries help people get up to speed quickly.`

	userPrompt := fmt.Sprintf(`Analyze the following comments from JIRA ticket %s and provide a summary in this exact format:

%s

Keep each answer concise - focus on the most important information that helps someone quickly understand the situation.
This summary will be added to a JIRA ticket as a comment. Please format the response using JIRA wiki markup (e.g., *bold*, _italic_, h3. headers).
Don't include a header for the summary, just the answers.

Please respond using this exact template format - don't change the headers:

h3. How long has the ticket been open?
[ Your answer here - check timestamps and calculate duration ]

h3. How many days of active research have SREs put into the ticket thus far?
[ Your answer here - count days with actual investigative work ]

h3. How many comments have been added to threads? Please add links.
[ Your answer here - look through all JIRA comments for Slack thread links (like https://redhat-internal.slack.com/archives/...) and list them. If no Slack links found, say "No Slack threads referenced." ]

h3. What actions have been taken?
[ Your answer here - list key troubleshooting steps, tests, changes made ]

h3. Summary of knowledge gained since the ticket opened
[ Your answer here - what has been learned, discovered, or ruled out ]

h3. What is the leading active hypothesis?
[ Your answer here - current theory about root cause or solution approach ]

h3. What do we not know that we need to find out next?
[ Your answer here - open questions, missing information, unknowns ]

h3. What are possible next steps and who is responsible for them?
[ Your answer here - specific actionable tasks and owners if mentioned ]
`, ticketKey, formattedComments)

	response, err := client.ChatCompletion(systemPrompt, userPrompt, s.modelName)
	if err != nil {
		return "", fmt.Errorf("AI API call failed: %w", err)
	}

	return response, nil
}
