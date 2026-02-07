package availability

type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

type TimeSlot struct {
	Start string `json:"start" bson:"start"` // "09:00"
	End   string `json:"end" bson:"end"`     // "17:00"
}

type DayAvailability struct {
	Day       DayOfWeek  `json:"day" bson:"day"`
	IsEnabled bool       `json:"is_enabled" bson:"is_enabled"`
	Slots     []TimeSlot `json:"slots" bson:"slots"`
}

type Availability struct {
	UserID              string            `json:"user_id" bson:"user_id"`
	Schedule            []DayAvailability `json:"schedule" bson:"schedule"`
	Timezone            string            `json:"timezone" bson:"timezone"`
	IsAcceptingBookings bool              `json:"is_accepting_bookings" bson:"is_accepting_bookings"`
}

func DefaultAvailability(userID string) *Availability {
	schedule := make([]DayAvailability, 7)
	for i := 0; i < 7; i++ {
		isEnabled := i != int(Sunday) && i != int(Saturday) // Mon-Fri enabled by default
		schedule[i] = DayAvailability{
			Day:       DayOfWeek(i),
			IsEnabled: isEnabled,
			Slots: []TimeSlot{
				{Start: "09:00", End: "17:00"},
			},
		}
	}
	return &Availability{
		UserID:              userID,
		Schedule:            schedule,
		Timezone:            "UTC",
		IsAcceptingBookings: true,
	}
}
