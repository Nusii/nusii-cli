package integration

import (
	"os"
	"testing"

	"github.com/nusii/nusii-cli/internal/api"
	"github.com/nusii/nusii-cli/internal/auth"
	"github.com/nusii/nusii-cli/internal/models"
)

func skipUnlessIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("NUSII_INTEGRATION_TEST") != "1" {
		t.Skip("Set NUSII_INTEGRATION_TEST=1 to run integration tests")
	}
}

func newTestClient(t *testing.T) *api.Client {
	t.Helper()
	apiKey := os.Getenv("NUSII_API_KEY")
	if apiKey == "" {
		t.Fatal("NUSII_API_KEY must be set for integration tests")
	}
	apiURL := os.Getenv("NUSII_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:3000"
	}
	return api.NewClient(apiURL, auth.NewTokenAuth(apiKey))
}

func TestAccountIntegration(t *testing.T) {
	skipUnlessIntegration(t)
	client := newTestClient(t)

	_, result, err := client.GetAccount()
	if err != nil {
		t.Fatalf("GetAccount failed: %v", err)
	}
	if result.Data.Attributes.Email == "" {
		t.Error("expected non-empty email")
	}
	t.Logf("Account: %s (%s)", result.Data.Attributes.Name, result.Data.Attributes.Email)
}

func TestFullCRUDCycle(t *testing.T) {
	skipUnlessIntegration(t)
	client := newTestClient(t)

	// 1. Create client
	_, clientResult, err := client.CreateClient(models.Client{
		Name:    "Integration Test",
		Email:   "integration@test.com",
		Surname: "User",
	})
	if err != nil {
		t.Fatalf("CreateClient failed: %v", err)
	}
	clientID := clientResult.Data.ID
	t.Logf("Created client: %s", clientID)

	// 2. Get client
	_, getResult, err := client.GetClient(clientID)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if getResult.Data.Attributes.Name != "Integration Test" {
		t.Errorf("expected name 'Integration Test', got '%s'", getResult.Data.Attributes.Name)
	}

	// 3. Update client
	_, updateResult, err := client.UpdateClient(clientID, models.Client{Name: "Updated Test"})
	if err != nil {
		t.Fatalf("UpdateClient failed: %v", err)
	}
	if updateResult.Data.Attributes.Name != "Updated Test" {
		t.Errorf("expected updated name, got '%s'", updateResult.Data.Attributes.Name)
	}

	// 4. List clients
	_, listResult, err := client.ListClients(1, 10)
	if err != nil {
		t.Fatalf("ListClients failed: %v", err)
	}
	if len(listResult.Data) == 0 {
		t.Error("expected at least one client")
	}

	// 5. Create proposal
	_, proposalResult, err := client.CreateProposal(models.Proposal{
		Title:    "Integration Test Proposal",
		ClientID: clientResult.Data.Attributes.ID,
	})
	if err != nil {
		t.Fatalf("CreateProposal failed: %v", err)
	}
	proposalID := proposalResult.Data.ID
	t.Logf("Created proposal: %s", proposalID)

	// 6. Create section
	_, sectionResult, err := client.CreateSection(models.Section{
		ProposalID:  proposalResult.Data.Attributes.ID,
		Title:       "Test Section",
		SectionType: "cost",
	})
	if err != nil {
		t.Fatalf("CreateSection failed: %v", err)
	}
	sectionID := sectionResult.Data.ID
	t.Logf("Created section: %s", sectionID)

	// 7. Create line item
	_, lineItemResult, err := client.CreateLineItem(sectionID, models.LineItem{
		Name:          "Test Item",
		Quantity:      2,
		AmountInCents: 10000,
	})
	if err != nil {
		t.Fatalf("CreateLineItem failed: %v", err)
	}
	lineItemID := lineItemResult.Data.ID
	t.Logf("Created line item: %s", lineItemID)

	// 8. List sections
	_, sectionsResult, err := client.ListSections(0, 0, proposalID, "", false)
	if err != nil {
		t.Fatalf("ListSections failed: %v", err)
	}
	if len(sectionsResult.Data) == 0 {
		t.Error("expected at least one section")
	}

	// 9. List line items
	_, lineItemsResult, err := client.ListLineItems(0, 0, sectionID)
	if err != nil {
		t.Fatalf("ListLineItems failed: %v", err)
	}
	if len(lineItemsResult.Data) == 0 {
		t.Error("expected at least one line item")
	}

	// Cleanup: delete in reverse order
	if err := client.DeleteLineItem(lineItemID); err != nil {
		t.Logf("Warning: DeleteLineItem failed: %v", err)
	}
	if err := client.DeleteSection(sectionID); err != nil {
		t.Logf("Warning: DeleteSection failed: %v", err)
	}
	if err := client.DeleteProposal(proposalID); err != nil {
		t.Logf("Warning: DeleteProposal failed: %v", err)
	}
	if err := client.DeleteClient(clientID); err != nil {
		t.Logf("Warning: DeleteClient failed: %v", err)
	}
	t.Log("Cleanup complete")
}

func TestListThemesIntegration(t *testing.T) {
	skipUnlessIntegration(t)
	client := newTestClient(t)

	_, result, err := client.ListThemes()
	if err != nil {
		t.Fatalf("ListThemes failed: %v", err)
	}
	t.Logf("Found %d themes", len(result.Data))
}

func TestListUsersIntegration(t *testing.T) {
	skipUnlessIntegration(t)
	client := newTestClient(t)

	_, result, err := client.ListUsers(0, 0)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(result.Data) == 0 {
		t.Error("expected at least one user")
	}
	t.Logf("Found %d users", len(result.Data))
}
