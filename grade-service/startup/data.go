package startup

import (
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"time"
)

var reviews = []*domain.Review{
	{
		Comment:            "Luxury Villa",
		Grade:              2.5,
		SubReviewer:        "66573d8fb73585ebf9ae0751",
		SubReviewed:        "57325353-5469-4930-8ec9-35c003e1b967",
		ReviewerFullName:   "Zorica Vukovic",
		DateOfModification: time.Now(),
	},
	{
		Comment:            "At least everything was new. Apartment was clean. Excellent accommodation!",
		Grade:              4,
		SubReviewer:        "66573d8fb73585ebf9ae0751",
		SubReviewed:        "57325353-5469-4930-8ec9-35c003e1b967",
		ReviewerFullName:   "Saska Topalovic",
		DateOfModification: time.Now(),
	},
	{
		Comment:            "Luxury Villa 2",
		Grade:              4,
		SubReviewer:        "66573d8fb73585ebf9ae0751",
		SubReviewed:        "88895353-5469-4930-8ec9-35c003e1b967",
		ReviewerFullName:   "Saska Topalovic",
		DateOfModification: time.Now(),
	},
}
