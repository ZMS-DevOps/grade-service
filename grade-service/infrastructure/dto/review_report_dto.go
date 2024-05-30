package dto

type NumberOfStars struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

func GetDefaultNumberOfStars() []NumberOfStars {
	return []NumberOfStars{
		{"1", 0},
		{"2", 0},
		{"3", 0},
		{"4", 0},
		{"5", 0},
	}
}

type ReviewReportDTO struct {
	TotalReviews  int             `json:"totalReviews"`
	AverageRating float32         `json:"averageRating"`
	NumberOfStars []NumberOfStars `json:"numberOfStars"`
	Reviews       []ReviewDTO     `json:"reviews"`
}
