package models

type Client struct {
	ID          int    `json:"id,omitempty"`
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	Surname     string `json:"surname,omitempty"`
	FullName    string `json:"full_name,omitempty"`
	Currency    string `json:"currency,omitempty"`
	Business    string `json:"business,omitempty"`
	Locale      string `json:"locale,omitempty"`
	PDFPageSize string `json:"pdf_page_size,omitempty"`
	Web         string `json:"web,omitempty"`
	Telephone   string `json:"telephone,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	Country     string `json:"country,omitempty"`
	State       string `json:"state,omitempty"`
}

func ClientTableHeaders() []string {
	return []string{"ID", "Name", "Email", "Business", "Currency"}
}

func (c Client) ClientTableRow(id string) []string {
	name := c.FullName
	if name == "" {
		name = c.Name
	}
	return []string{id, name, c.Email, c.Business, c.Currency}
}
