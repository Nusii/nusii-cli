package models

import "strings"

type WebhookEndpoint struct {
	ID        int      `json:"id,omitempty"`
	TargetURL string   `json:"target_url,omitempty"`
	Events    []string `json:"events,omitempty"`
}

func WebhookTableHeaders() []string {
	return []string{"ID", "Target URL", "Events"}
}

func (w WebhookEndpoint) WebhookTableRow(id string) []string {
	return []string{id, w.TargetURL, strings.Join(w.Events, ", ")}
}
