package entity

// UserSettings represents settings stored per user in Redis (hash fields)
type UserSettings struct {
	Theme       string `json:"theme"`
	PickedModel string `json:"picked_model"`
	Font        string `json:"font"`
}
