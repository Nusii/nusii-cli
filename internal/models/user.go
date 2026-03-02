package models

type User struct {
	ID    int    `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func UserTableHeaders() []string {
	return []string{"ID", "Name", "Email"}
}

func (u User) UserTableRow(id string) []string {
	return []string{id, u.Name, u.Email}
}
