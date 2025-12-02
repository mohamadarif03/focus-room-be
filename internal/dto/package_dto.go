package dto

import "time"

type PackageRequest struct {
	Title     string `json:"title" binding:"required"`
	ColorIcon string `json:"color_icon" binding:"required"`
}

type PackageResponse struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	ColorIcon     string    `json:"colorIcon"`
	UserID        uint      `json:"user_id"`
	MaterialCount int       `json:"material_count"`
	CreatedAt     time.Time `json:"created_at"`
}
type AdminMaterialRequest struct {
	Title      string `json:"title" binding:"required"`
	SourceType string `json:"source_type" binding:"required,oneof=youtube pdf link"`
	Source     string `json:"source" binding:"required"`
}

type AdminCreatePackageRequest struct {
	Title     string                 `json:"title" binding:"required"`
	ColorIcon string                 `json:"color_icon" binding:"required"`
	IsPublic  bool                   `json:"is_public"`
	Materials []AdminMaterialRequest `json:"materials" binding:"required,dive"`
}
type PackageWithMaterialsResponse struct {
	ID        uint             `json:"id"`
	Title     string           `json:"title"`
	ColorIcon string           `json:"colorIcon"`
	Materials []MaterialSimple `json:"materials"`
}

type MaterialSimple struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	SourceType string    `json:"source_type"`
	CreatedAt  time.Time `json:"created_at"`
}
