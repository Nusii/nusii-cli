package models

type Proposal struct {
	ID                    int    `json:"id,omitempty"`
	Title                 string `json:"title,omitempty"`
	AccountID             int    `json:"account_id,omitempty"`
	Status                string `json:"status,omitempty"`
	PublicID              string `json:"public_id,omitempty"`
	DocumentNumber        string `json:"document_number,omitempty"`
	ClientID              int    `json:"client_id,omitempty"`
	ClientEmail           string `json:"client_email,omitempty"`
	SenderID              int    `json:"sender_id,omitempty"`
	PreparedByID          int    `json:"prepared_by_id,omitempty"`
	Currency              string `json:"currency,omitempty"`
	Theme                 string `json:"theme,omitempty"`
	SentAt                string `json:"sent_at,omitempty"`
	AcceptedAt            string `json:"accepted_at,omitempty"`
	ArchivedAt            string `json:"archived_at,omitempty"`
	ExpiresAt             string `json:"expires_at,omitempty"`
	DisplayDate           string `json:"display_date,omitempty"`
	Report                bool   `json:"report,omitempty"`
	ExcludeTotal          bool   `json:"exclude_total,omitempty"`
	ExcludeTotalInPDF     bool   `json:"exclude_total_in_pdf,omitempty"`
	AcceptedTotalInCents  int    `json:"accepted_total_in_cents,omitempty"`
	AcceptedTotalFormatted string `json:"accepted_total_formatted,omitempty"`
	SenderName            string `json:"sender_name,omitempty"`
	Source                string `json:"source,omitempty"`
	TemplateID            int    `json:"template_id,omitempty"`
}

func ProposalTableHeaders() []string {
	return []string{"ID", "Title", "Status", "Client ID", "Currency", "Sent At"}
}

func (p Proposal) ProposalTableRow(id string) []string {
	return []string{id, p.Title, p.Status, itoa(p.ClientID), p.Currency, p.SentAt}
}

// ProposalSendRequest represents the payload for sending a proposal.
type ProposalSendRequest struct {
	Email   string `json:"email,omitempty"`
	CC      string `json:"cc,omitempty"`
	BCC     string `json:"bcc,omitempty"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
}
