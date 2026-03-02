package models

type LineItem struct {
	ID              int    `json:"id,omitempty"`
	SectionID       int    `json:"section_id,omitempty"`
	Name            string `json:"name,omitempty"`
	Position        int    `json:"position,omitempty"`
	CostType        string `json:"cost_type,omitempty"`
	RecurringType   string `json:"recurring_type,omitempty"`
	PerType         string `json:"per_type,omitempty"`
	Quantity        int    `json:"quantity,omitempty"`
	Currency        string `json:"currency,omitempty"`
	AmountInCents   int    `json:"amount_in_cents,omitempty"`
	AmountFormatted string `json:"amount_formatted,omitempty"`
	TotalInCents    int    `json:"total_in_cents,omitempty"`
	TotalFormatted  string `json:"total_formatted,omitempty"`
}

func LineItemTableHeaders() []string {
	return []string{"ID", "Name", "Section ID", "Quantity", "Amount", "Total"}
}

func (l LineItem) LineItemTableRow(id string) []string {
	return []string{id, l.Name, itoa(l.SectionID), itoa(l.Quantity), l.AmountFormatted, l.TotalFormatted}
}
