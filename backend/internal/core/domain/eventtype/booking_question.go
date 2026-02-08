package eventtype

type QuestionType string

const (
	QuestionTypeText   QuestionType = "text"
	QuestionTypePhone  QuestionType = "phone"
	QuestionTypeSelect QuestionType = "select"
)

type BookingQuestion struct {
	ID       string       `json:"id" bson:"id"`
	Label    string       `json:"label" bson:"label"`
	Type     QuestionType `json:"type" bson:"type"`
	Required bool         `json:"required" bson:"required"`
	Options  []string     `json:"options,omitempty" bson:"options,omitempty"` // For select type
}
