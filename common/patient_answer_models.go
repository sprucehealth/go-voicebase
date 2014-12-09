package common

import (
	"time"

	"github.com/sprucehealth/backend/encoding"
)

type Answer interface {
	getQuestionId() int64
}

type AnswerIntake struct {
	AnswerIntakeID    encoding.ObjectID `json:"answer_id,omitempty"`
	QuestionID        encoding.ObjectID `json:"question_id,omitempty"`
	RoleID            encoding.ObjectID `json:"-"`
	Role              string            `json:"-"`
	ContextId         encoding.ObjectID `json:"-"`
	ParentQuestionId  encoding.ObjectID `json:"-"`
	ParentAnswerId    encoding.ObjectID `json:"-"`
	PotentialAnswerID encoding.ObjectID `json:"potential_answer_id,omitempty"`
	PotentialAnswer   string            `json:"potential_answer,omitempty"`
	AnswerSummary     string            `json:"potential_answer_summary,omitempty"`
	LayoutVersionID   encoding.ObjectID `json:"-"`
	SubAnswers        []*AnswerIntake   `json:"answers,omitempty"`
	AnswerText        string            `json:"answer_text,omitempty"`
	ObjectUrl         string            `json:"object_url,omitempty"`
	StorageBucket     string            `json:"-"`
	StorageKey        string            `json:"-"`
	StorageRegion     string            `json:"-"`
	ToAlert           bool              `json:"-"`
}

func (a *AnswerIntake) getQuestionId() int64 {
	return a.QuestionID.Int64()
}

type PhotoIntakeSection struct {
	ID           int64              `json:"-"`
	QuestionID   int64              `json:"-"`
	Name         string             `json:"name,omitempty"`
	Photos       []*PhotoIntakeSlot `json:"photos,omitempty"`
	CreationDate time.Time          `json:"creation_date"`
}

func (p *PhotoIntakeSection) getQuestionId() int64 {
	return p.QuestionID
}

type PhotoIntakeSlot struct {
	ID           int64     `json:"-"`
	CreationDate time.Time `json:"creation_date"`
	PhotoURL     string    `json:"photo_url"`
	PhotoID      int64     `json:"photo_id,string,omitempty"`
	SlotID       int64     `json:"slot_id,string"`
	Name         string    `json:"name"`
}
