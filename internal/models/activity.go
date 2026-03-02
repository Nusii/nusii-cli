package models

type Activity struct {
	ID                 int    `json:"id,omitempty"`
	ActivityType       string `json:"activity_type,omitempty"`
	IPAddress          string `json:"ip_address,omitempty"`
	AdditionalFields   any    `json:"additional_fields,omitempty"`
	ProposalTitle      string `json:"proposal_title,omitempty"`
	ProposalCreatedAt  string `json:"proposal_created_at,omitempty"`
	ProposalSentAt     string `json:"proposal_sent_at,omitempty"`
	ProposalStatus     string `json:"proposal_status,omitempty"`
	ProposalPublicID   string `json:"proposal_public_id,omitempty"`
	ProposalExpiresAt  string `json:"proposal_expires_at,omitempty"`
	ClientName         string `json:"client_name,omitempty"`
	ClientEmail        string `json:"client_email,omitempty"`
	ClientSurname      string `json:"client_surname,omitempty"`
	ClientFullName     string `json:"client_full_name,omitempty"`
	ClientBusiness     string `json:"client_business,omitempty"`
	ClientTelephone    string `json:"client_telephone,omitempty"`
	ClientLocale       string `json:"client_locale,omitempty"`
}

func ActivityTableHeaders() []string {
	return []string{"ID", "Type", "Proposal Title", "Client", "Status"}
}

func (a Activity) ActivityTableRow(id string) []string {
	client := a.ClientFullName
	if client == "" {
		client = a.ClientName
	}
	return []string{id, a.ActivityType, a.ProposalTitle, client, a.ProposalStatus}
}
