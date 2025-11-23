package model

import (
	"time"

	"gorm.io/datatypes"
)


type Quiz struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	MaterialID uint `gorm:"not null" json:"material_id"`
	CreatedAt  time.Time

	Material  Material       `gorm:"foreignKey:MaterialID"`
	Questions []QuizQuestion `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE;" json:"questions"`
}

type QuizQuestion struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	QuizID       uint           `gorm:"not null" json:"quiz_id"`
	Pertanyaan   string         `gorm:"type:text;not null" json:"pertanyaan"`
	Pilihan      datatypes.JSON `gorm:"type:json" json:"pilihan"`
	JawabanBenar string         `gorm:"size:5" json:"jawaban_benar"`
}


type QuizAttempt struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	QuizID    uint      `gorm:"not null" json:"quiz_id"`
	Score     int       `gorm:"not null" json:"score"` 
	CreatedAt time.Time `json:"created_at"`

	User    User                `gorm:"foreignKey:UserID"`
	Quiz    Quiz                `gorm:"foreignKey:QuizID"`
	Answers []QuizAttemptDetail `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE;" json:"answers"`
}

type QuizAttemptDetail struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	AttemptID  uint   `gorm:"not null" json:"attempt_id"`
	QuestionID uint   `gorm:"not null" json:"question_id"`
	UserAnswer string `gorm:"size:5" json:"user_answer"` 
	IsCorrect  bool   `json:"is_correct"`               
}
