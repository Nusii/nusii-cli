package models

type Account struct {
	ID           int    `json:"id,omitempty"`
	Email        string `json:"email,omitempty"`
	Name         string `json:"name,omitempty"`
	Subdomain    string `json:"subdomain,omitempty"`
	Web          string `json:"web,omitempty"`
	Currency     string `json:"currency,omitempty"`
	PDFPageSize  string `json:"pdf_page_size,omitempty"`
	Locale       string `json:"locale,omitempty"`
	Address      string `json:"address,omitempty"`
	AddressState string `json:"address_state,omitempty"`
	Postcode     string `json:"postcode,omitempty"`
	City         string `json:"city,omitempty"`
	Telephone    string `json:"telephone,omitempty"`
	DefaultTheme string `json:"default_theme,omitempty"`
}

// AccountTableHeaders returns column headers for table output.
func AccountTableHeaders() []string {
	return []string{"ID", "Name", "Email", "Subdomain", "Currency", "Locale"}
}

// AccountTableRow returns a row for table output.
func (a Account) AccountTableRow(id string) []string {
	return []string{id, a.Name, a.Email, a.Subdomain, a.Currency, a.Locale}
}
