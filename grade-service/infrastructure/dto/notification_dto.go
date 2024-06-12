package dto

type NotificationDTO struct {
	UserId          string `json:"receiver_id"`
	ReviewerName    string `json:"start_action_user_name"`
	AccommodationId string `json:"accommodation_id"`
}
