package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"governance-alerts-cosmos/internal/types"
)

// Client represents a governance client
type Client struct {
	config types.NetworkConfig
	client *http.Client
}

// CosmosGovResponse represents the response from Cosmos governance API
type CosmosGovResponse struct {
	Proposals  []CosmosProposal `json:"proposals"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// CosmosProposal represents a proposal from Cosmos governance API
type CosmosProposal struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	VotingStart string `json:"voting_start_time"`
	VotingEnd   string `json:"voting_end_time"`
	Messages    []struct {
		TypeURL string `json:"@type"`
	} `json:"messages"`
}

// NewClient creates a new governance client
func NewClient(config types.NetworkConfig) (*Client, error) {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	return nil
}

// GetVotingProposals fetches all proposals and filters voting ones
func (c *Client) GetVotingProposals(ctx context.Context) ([]types.Proposal, error) {
	fmt.Printf("Checking proposals for %s (%s)\n", c.config.Name, c.config.ChainID)

	// Build API URL for all proposals
	apiURL := fmt.Sprintf("%s/cosmos/gov/v1/proposals", c.config.RestEndpoint)
	fmt.Printf("  API URL: %s\n", apiURL)

	// Make HTTP request
	body, err := c.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proposals: %w", err)
	}

	// Parse response
	var response CosmosGovResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("  Found %d total proposals\n", len(response.Proposals))

	// Filter proposals in voting period
	proposals := make([]types.Proposal, 0)
	for _, proposal := range response.Proposals {
		if proposal.Status == "PROPOSAL_STATUS_VOTING_PERIOD" {
			// Parse voting start time
			votingStart, err := time.Parse(time.RFC3339, proposal.VotingStart)
			if err != nil {
				fmt.Printf("Warning: failed to parse voting start time for proposal %s: %v\n", proposal.ID, err)
				continue
			}

			// Parse voting end time
			votingEnd, err := time.Parse(time.RFC3339, proposal.VotingEnd)
			if err != nil {
				fmt.Printf("Warning: failed to parse voting end time for proposal %s: %v\n", proposal.ID, err)
				continue
			}

			// Get proposal title and description
			title := proposal.Title
			if title == "" {
				title = fmt.Sprintf("Proposal %s", proposal.ID)
			}

			description := proposal.Description
			if description == "" {
				description = "No description available"
			}

			// Convert ID to uint64
			var proposalID uint64
			if _, err := fmt.Sscanf(proposal.ID, "%d", &proposalID); err != nil {
				fmt.Printf("Warning: failed to parse proposal ID %s: %v\n", proposal.ID, err)
				continue
			}

			proposals = append(proposals, types.Proposal{
				ID:          proposalID,
				Title:       title,
				Description: description,
				Status:      proposal.Status,
				VotingStart: votingStart,
				VotingEnd:   votingEnd,
				Network:     c.config.Name,
			})
		}
	}

	fmt.Printf("  Found %d proposals in voting period\n", len(proposals))
	return proposals, nil
}

// GetProposalDetails fetches detailed information about a specific proposal
func (c *Client) GetProposalDetails(ctx context.Context, proposalID uint64) (*types.Proposal, error) {
	// Build API URL for specific proposal
	apiURL := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%d", c.config.RestEndpoint, proposalID)

	// Make HTTP request
	body, err := c.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch proposal %d: %w", proposalID, err)
	}

	// Parse response
	var response struct {
		Proposal CosmosProposal `json:"proposal"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	proposal := response.Proposal

	// Parse voting start time
	votingStart, err := time.Parse(time.RFC3339, proposal.VotingStart)
	if err != nil {
		return nil, fmt.Errorf("failed to parse voting start time: %w", err)
	}

	// Parse voting end time
	votingEnd, err := time.Parse(time.RFC3339, proposal.VotingEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to parse voting end time: %w", err)
	}

	// Get proposal title and description
	title := proposal.Title
	if title == "" {
		title = fmt.Sprintf("Proposal %s", proposal.ID)
	}

	description := proposal.Description
	if description == "" {
		description = "No description available"
	}

	// Convert ID to uint64
	var id uint64
	if _, err := fmt.Sscanf(proposal.ID, "%d", &id); err != nil {
		return nil, fmt.Errorf("failed to parse proposal ID: %w", err)
	}

	return &types.Proposal{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      proposal.Status,
		VotingStart: votingStart,
		VotingEnd:   votingEnd,
		Network:     c.config.Name,
	}, nil
}

// CheckProposalStatus checks if a proposal is in voting period
func (c *Client) CheckProposalStatus(ctx context.Context, proposalID uint64) (string, error) {
	proposal, err := c.GetProposalDetails(ctx, proposalID)
	if err != nil {
		return "", err
	}
	return proposal.Status, nil
}

// Helper function to make HTTP requests
func (c *Client) makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Governance-Alerts-Cosmos/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
