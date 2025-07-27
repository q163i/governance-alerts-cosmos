package service

import (
	"context"
	"fmt"
	"time"

	"governance-alerts-cosmos/internal/governance"
	"governance-alerts-cosmos/internal/notifications"
	"governance-alerts-cosmos/internal/types"
)

// Service represents the governance alerts service
type Service struct {
	config   *types.Config
	notifier *notifications.Notifier
	clients  map[string]*governance.Client
	stopChan chan struct{}
}

// NewService creates a new governance alerts service
func NewService(config *types.Config) (*Service, error) {
	// Initialize notifier
	notifier, err := notifications.NewNotifier(&config.Notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifier: %w", err)
	}

	// Initialize governance clients for each network
	clients := make(map[string]*governance.Client)
	for name, networkConfig := range config.Networks {
		client, err := governance.NewClient(networkConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create client for %s: %w", name, err)
		}
		clients[name] = client
	}

	return &Service{
		config:   config,
		notifier: notifier,
		clients:  clients,
		stopChan: make(chan struct{}),
	}, nil
}

// Run starts the governance alerts service
func (s *Service) Run(ctx context.Context) error {
	// Send startup notification if enabled
	if s.config.Alerts.NotifyOnStartup {
		if err := s.sendStartupNotification(); err != nil {
			fmt.Printf("Warning: failed to send startup notification: %v\n", err)
		}
	}

	fmt.Println("Starting Governance Alerts Service...")

	// Start monitoring loop
	ticker := time.NewTicker(time.Duration(s.config.Alerts.CheckIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Initial check
	if err := s.checkProposals(ctx); err != nil {
		fmt.Printf("Error during initial check: %v\n", err)
	}

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			return nil
		case <-ticker.C:
			if err := s.checkProposals(ctx); err != nil {
				fmt.Printf("Error checking proposals: %v\n", err)
			}
		}
	}
}

// Stop stops the service
func (s *Service) Stop() {
	close(s.stopChan)
}

// sendStartupNotification sends a notification when the service starts
func (s *Service) sendStartupNotification() error {
	networks := make([]string, 0, len(s.config.Networks))
	for _, network := range s.config.Networks {
		networks = append(networks, fmt.Sprintf("%s (%s)", network.Name, network.ChainID))
	}

	msg := types.NotificationMessage{
		Title:       "🚀 Governance Alerts Service Started",
		Content:     fmt.Sprintf("Service is now monitoring %d networks:\n• %s", len(networks), networks[0]),
		Network:     "Governance Alerts",
		ChainID:     "Service",
		ProposalID:  0,
		ExplorerURL: "",
	}

	// Add additional networks if more than one
	if len(networks) > 1 {
		for i := 1; i < len(networks); i++ {
			msg.Content += fmt.Sprintf("\n• %s", networks[i])
		}
	}

	return s.notifier.SendNotification(msg)
}

// checkProposals checks all networks for proposals
func (s *Service) checkProposals(ctx context.Context) error {
	fmt.Printf("Checking proposals at %s\n", time.Now().Format(time.RFC3339))

	for name, client := range s.clients {
		if err := s.checkNetworkProposals(ctx, name, client); err != nil {
			fmt.Printf("Error checking proposals for %s: %v\n", name, err)
		}
	}

	return nil
}

// checkNetworkProposals checks proposals for a specific network
func (s *Service) checkNetworkProposals(ctx context.Context, networkName string, client *governance.Client) error {
	proposals, err := client.GetVotingProposals(ctx)
	if err != nil {
		return fmt.Errorf("failed to get proposals: %w", err)
	}

	if len(proposals) == 0 {
		fmt.Printf("  No active proposals found for %s\n", networkName)
		return nil
	}

	fmt.Printf("  Found %d active proposals for %s\n", len(proposals), networkName)

	networkConfig := s.config.Networks[networkName]
	for _, proposal := range proposals {
		if err := s.checkProposal(ctx, proposal, client, networkConfig); err != nil {
			fmt.Printf("Error checking proposal %d: %v\n", proposal.ID, err)
		}
	}

	return nil
}

// checkProposal checks a specific proposal and sends notifications if needed
func (s *Service) checkProposal(ctx context.Context, proposal types.Proposal, client *governance.Client, networkConfig types.NetworkConfig) error {
	now := time.Now()

	// Log proposal details
	fmt.Printf("  📋 Proposal %d: %s\n", proposal.ID, proposal.Title)
	fmt.Printf("     Description: %s\n", truncateString(proposal.Description, 100))
	fmt.Printf("     Network: %s (%s)\n", proposal.Network, networkConfig.ChainID)
	fmt.Printf("     Voting: %s → %s\n",
		proposal.VotingStart.Format("2006-01-02 15:04:05"),
		proposal.VotingEnd.Format("2006-01-02 15:04:05"))

	// Check if we should notify about voting start
	if proposal.VotingStart.After(now) {
		timeUntilStart := proposal.VotingStart.Sub(now)
		hoursUntilStart := timeUntilStart.Hours()

		if hoursUntilStart <= float64(s.config.Alerts.HoursBeforeStart) && hoursUntilStart > 0 {
			msg := types.NotificationMessage{
				Title:       fmt.Sprintf("🚨 Governance Proposal Voting Starting Soon - %s", proposal.Network),
				Content:     fmt.Sprintf("Proposal \"%s\" will start voting in %.1f hours.\n\nDescription: %s", proposal.Title, hoursUntilStart, proposal.Description),
				Network:     proposal.Network,
				ChainID:     networkConfig.ChainID,
				ProposalID:  proposal.ID,
				ExplorerURL: "",
			}

			if err := s.notifier.SendNotification(msg); err != nil {
				return fmt.Errorf("failed to send start notification: %w", err)
			}

			fmt.Printf("     ✅ Sent start notification (%.1f hours until start)\n", hoursUntilStart)
		} else {
			fmt.Printf("     ⏰ Start notification not needed (%.1f hours until start)\n", hoursUntilStart)
		}
	}

	// Check if we should notify about voting end
	if proposal.VotingEnd.After(now) {
		timeUntilEnd := proposal.VotingEnd.Sub(now)
		hoursUntilEnd := timeUntilEnd.Hours()

		if hoursUntilEnd <= float64(s.config.Alerts.HoursBeforeEnd) && hoursUntilEnd > 0 {
			msg := types.NotificationMessage{
				Title:       fmt.Sprintf("⏰ Governance Proposal Voting Ending Soon - %s", proposal.Network),
				Content:     fmt.Sprintf("Proposal \"%s\" will end voting in %.1f hours.\n\nDescription: %s", proposal.Title, hoursUntilEnd, proposal.Description),
				Network:     proposal.Network,
				ChainID:     networkConfig.ChainID,
				ProposalID:  proposal.ID,
				ExplorerURL: "",
			}

			if err := s.notifier.SendNotification(msg); err != nil {
				return fmt.Errorf("failed to send end notification: %w", err)
			}

			fmt.Printf("     ✅ Sent end notification (%.1f hours until end)\n", hoursUntilEnd)
		} else {
			fmt.Printf("     ⏰ End notification not needed (%.1f hours until end)\n", hoursUntilEnd)
		}
	}

	fmt.Printf("     ---\n")
	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
