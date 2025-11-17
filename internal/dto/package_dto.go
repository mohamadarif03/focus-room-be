package dto

import "time"

type PackageRequest struct {
	Title     string `json:"title" binding:"required"`
	ColorIcon string `json:"color_icon" binding:"required"`
}

type PackageResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	ColorIcon string    `json:"colorIcon"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PackageWithMaterialsResponse struct {
	ID        uint             `json:"id"`
	Title     string           `json:"title"`
	ColorIcon string           `json:"colorIcon"`
	Materials []MaterialSimple `json:"materials"`
}

type MaterialSimple struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	SourceType string `json:"source_type"`
}
