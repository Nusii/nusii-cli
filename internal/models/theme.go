package models

type Theme struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func ThemeTableHeaders() []string {
	return []string{"ID", "Name"}
}

func (t Theme) ThemeTableRow(id string) []string {
	return []string{id, t.Name}
}
