package schemas

type HardwareCreate struct {
	HardwareName string `json:"hardwareName" validate:"required,min=1,max=200"`
	Capacity     int    `json:"capacity" validate:"required,gte=1"`
}

type HardwareUpdate struct {
	HardwareName *string `json:"hardwareName" validate:"omitempty,min=1,max=200"`
	Capacity     *int    `json:"capacity" validate:"omitempty,gte=1"`
}

type HardwareCheckout struct {
	ProjectID string `json:"projectId" validate:"required,min=1"`
	Amount    int    `json:"amount" validate:"required,gte=1"`
	UserID    string `json:"userId" validate:"required,min=1"`
}

type HardwareCheckin struct {
	ProjectID string `json:"projectId" validate:"required,min=1"`
	Amount    int    `json:"amount" validate:"required,gte=1"`
	UserID    string `json:"userId" validate:"required,min=1"`
}
