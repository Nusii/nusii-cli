package models

type Section struct {
	ID             int    `json:"id,omitempty"`
	Currency       string `json:"currency,omitempty"`
	AccountID      int    `json:"account_id,omitempty"`
	ProposalID     int    `json:"proposal_id,omitempty"`
	TemplateID     int    `json:"template_id,omitempty"`
	Title          string `json:"title,omitempty"`
	Name           string `json:"name,omitempty"`
	Body           string `json:"body,omitempty"`
	Position       int    `json:"position,omitempty"`
	Reusable       bool   `json:"reusable,omitempty"`
	SectionType    string `json:"section_type,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	PageBreak      bool   `json:"page_break,omitempty"`
	Optional       bool   `json:"optional,omitempty"`
	Selected       bool   `json:"selected,omitempty"`
	IncludeTotal   bool   `json:"include_total,omitempty"`
	TotalInCents   int    `json:"total_in_cents,omitempty"`
	TotalFormatted string `json:"total_formatted,omitempty"`
}

func SectionTableHeaders() []string {
	return []string{"ID", "Title", "Proposal ID", "Type", "Position", "Total"}
}

func (s Section) SectionTableRow(id string) []string {
	return []string{id, s.Title, itoa(s.ProposalID), s.SectionType, itoa(s.Position), s.TotalFormatted}
}
