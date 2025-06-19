package models

import (
	"time"
)

// InvoiceLineItem represents a single line item in an invoice
type InvoiceLineItem struct {
	LineItemID     string     `json:"line_item_id"`
	InvoiceID      string     `json:"invoice_id"` // Foreign key
	Description    string     `json:"description"`
	Quantity       float64    `json:"quantity"`
	UnitPrice      float64    `json:"unit_price"`
	Amount         float64    `json:"amount"` // Calculated: Quantity * UnitPrice
	RelatedJobType *string    `json:"related_job_type"` // Pointer for nullable string
	RelatedJobID   *string    `json:"related_job_id"`   // Pointer for nullable UUID string
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Invoice represents an invoice document
type Invoice struct {
	InvoiceID         string            `json:"invoice_id"`
	InvoiceNumber     string            `json:"invoice_number"`
	ClientID          string            `json:"client_id"`
	ClientName        string            `json:"client_name"` // Populated from clients table
	NpwpClient        string            `json:"npwp_client"` // Populated from clients table
	AssignedStaffID   *string           `json:"assigned_staff_id"` // Pointer for nullable UUID string
	AssignedStaffName *string           `json:"assigned_staff_name"` // Populated from staffs table
	InvoiceDate     	CustomDate 				`gorm:"type:date;not null" json:"invoice_date"` // <-- UBAH DI SINI
  DueDate         	CustomDate 				`gorm:"type:date;not null" json:"due_date"`
	TotalAmount       float64           `json:"total_amount"`
	Status            string            `json:"status"` // e.g., "Draft", "Issued", "Paid"
	Notes             *string           `json:"notes"` // Pointer for nullable string
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	LineItems         []InvoiceLineItem `json:"line_items"` // Nested slice of line items
}

// NewInvoiceLineItemRequest for creating a line item
type NewInvoiceLineItemRequest struct {
	Description    string  `json:"description" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice      float64 `json:"unit_price" binding:"required,gt=0"`
	RelatedJobType *string `json:"related_job_type"`
	RelatedJobID   *string `json:"related_job_id"`
}

// NewInvoiceRequest for creating a new invoice
type NewInvoiceRequest struct {
	ClientID          string                      `json:"client_id" binding:"required"`
	AssignedStaffID   *string                     `json:"assigned_staff_id"`
	InvoiceDate       time.Time                   `json:"invoice_date" binding:"required"`
	DueDate           time.Time                   `json:"due_date" binding:"required"`
	Status            string                      `json:"status"`
	Notes             *string                     `json:"notes"`
	LineItems         []NewInvoiceLineItemRequest `json:"line_items" binding:"required,min=1"` // Min 1 line item
}

// UpdateInvoiceLineItemRequest for updating a line item
type UpdateInvoiceLineItemRequest struct {
	Description    *string  `json:"description"`
	Quantity       *float64 `json:"quantity"`
	UnitPrice      *float64 `json:"unit_price"`
	RelatedJobType *string  `json:"related_job_type"`
	RelatedJobID   *string  `json:"related_job_id"`
}

// UpdateInvoiceRequest for updating an invoice header
type UpdateInvoiceRequest struct {
	ClientID          *string    `json:"client_id"`
	AssignedStaffID   *string    `json:"assigned_staff_id"`
	InvoiceDate       *time.Time `json:"invoice_date"`
	DueDate           *time.Time `json:"due_date"`
	Status            *string    `json:"status"`
	Notes             *string    `json:"notes"`
}
