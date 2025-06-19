package repositories

import (
	"database/sql"
	"fmt" // Untuk string manipulation (invoice number)
	"time"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/models" // Pastikan ini adalah modul Anda
)

// InvoiceRepository defines the interface for invoice data operations
type InvoiceRepository interface {
	CreateInvoice(invoice *models.Invoice) error
	GetAllInvoices(staffIDFilter string, isAdmin bool) ([]models.Invoice, error)
	GetInvoiceByID(id string, staffIDFilter string, isAdmin bool) (*models.Invoice, error)
	UpdateInvoice(invoice *models.Invoice) error // Update header info
	DeleteInvoice(id string) error

	// Line item specific operations (if needed for individual line item management)
	CreateInvoiceLineItem(item *models.InvoiceLineItem) error
	UpdateInvoiceLineItem(item *models.InvoiceLineItem) error
	DeleteInvoiceLineItem(itemID string) error
}

// invoiceRepository implements InvoiceRepository interface
type invoiceRepository struct {
	db *sql.DB
}

// NewInvoiceRepository creates a new InvoiceRepository
func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

// generateInvoiceNumber generates a unique invoice number (e.g., INV/YYYYMMDD/SEQ)
func (r *invoiceRepository) generateInvoiceNumber(tx *sql.Tx, invoiceDate time.Time) (string, error) {
	datePrefix := invoiceDate.Format("20060102") // YYYYMMDD
	
	// Get the last sequence number for today
	var lastSeq int
	// Query to find the maximum sequence number for today's date
	seqQuery := `SELECT COALESCE(MAX(SUBSTRING(invoice_number FROM 11 FOR LENGTH(invoice_number))::INT), 0)
                 FROM invoices WHERE invoice_number LIKE $1 || '%' AND invoice_date = $2`
	
	err := tx.QueryRow(seqQuery, "INV/"+datePrefix, invoiceDate).Scan(&lastSeq)
	if err != nil && err != sql.ErrNoRows { // sql.ErrNoRows means no invoices yet for today, which is fine
		return "", fmt.Errorf("failed to get last invoice sequence: %w", err)
	}

	newSeq := lastSeq + 1
	return fmt.Sprintf("INV/%s/%03d", datePrefix, newSeq), nil // INV/YYYYMMDD/001
}

// calculateTotalAmount calculates the total amount from line items
func (r *invoiceRepository) calculateTotalAmount(lineItems []models.InvoiceLineItem) float64 {
	var total float64
	for _, item := range lineItems {
		total += item.Amount
	}
	return total
}

// CreateInvoice inserts a new invoice and its associated line items into the database.
// This operation is wrapped in a a transaction.
func (r *invoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Generate Invoice Number
	invoice.InvoiceNumber, err = r.generateInvoiceNumber(tx, invoice.InvoiceDate.Time)
	if err != nil {
		return fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// 2. Calculate Total Amount
	invoice.TotalAmount = r.calculateTotalAmount(invoice.LineItems)

	// 3. Handle nullable foreign keys and notes
	var assignedStaffID sql.NullString
	if invoice.AssignedStaffID != nil && *invoice.AssignedStaffID != "" {
		assignedStaffID = sql.NullString{String: *invoice.AssignedStaffID, Valid: true}
	} else {
		assignedStaffID = sql.NullString{Valid: false}
	}

	var notes sql.NullString
	if invoice.Notes != nil && *invoice.Notes != "" {
		notes = sql.NullString{String: *invoice.Notes, Valid: true}
	} else {
		notes = sql.NullString{Valid: false}
	}

	// 4. Insert into invoices table
	invoiceQuery := `INSERT INTO invoices (
		invoice_number, client_id, assigned_staff_id, invoice_date, due_date, total_amount, status, notes, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
	) RETURNING invoice_id, created_at, updated_at`

	if invoice.CreatedAt.IsZero() {
		invoice.CreatedAt = time.Now()
	}
	invoice.UpdatedAt = time.Now()

	err = tx.QueryRow(invoiceQuery,
		invoice.InvoiceNumber, invoice.ClientID, assignedStaffID, invoice.InvoiceDate, invoice.DueDate,
		invoice.TotalAmount, invoice.Status, notes,
		invoice.CreatedAt, invoice.UpdatedAt,
	).Scan(&invoice.InvoiceID, &invoice.CreatedAt, &invoice.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// 5. Insert associated line items
	if len(invoice.LineItems) > 0 {
		lineItemQuery := `INSERT INTO invoice_line_items (
			invoice_id, description, quantity, unit_price, amount, related_job_type, related_job_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING line_item_id, created_at, updated_at`

		for i := range invoice.LineItems {
			item := &invoice.LineItems[i]
			item.InvoiceID = invoice.InvoiceID

			// Handle nullable related_job fields
			var relatedJobType sql.NullString
			if item.RelatedJobType != nil && *item.RelatedJobType != "" {
				relatedJobType = sql.NullString{String: *item.RelatedJobType, Valid: true}
			} else {
				relatedJobType = sql.NullString{Valid: false}
			}

			var relatedJobID sql.NullString
			if item.RelatedJobID != nil && *item.RelatedJobID != "" {
				relatedJobID = sql.NullString{String: *item.RelatedJobID, Valid: true}
			} else {
				relatedJobID = sql.NullString{Valid: false}
			}

			if item.CreatedAt.IsZero() {
				item.CreatedAt = time.Now()
			}
			item.UpdatedAt = time.Now()

			err := tx.QueryRow(lineItemQuery,
				item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.Amount,
				relatedJobType, relatedJobID,
				item.CreatedAt, item.UpdatedAt,
			).Scan(&item.LineItemID, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create invoice line item: %w", err)
			}
		}
	}

	return tx.Commit()
}

// GetAllInvoices fetches all invoices with their associated client and staff info.
func (r *invoiceRepository) GetAllInvoices(staffIDFilter string, isAdmin bool) ([]models.Invoice, error) {
	query := `
	SELECT
		i.invoice_id, i.invoice_number, i.client_id, c.client_name, c.npwp_client,
		i.assigned_staff_id, s.nama AS assigned_staff_name,
		i.invoice_date, i.due_date, i.total_amount, i.status, i.notes,
		i.created_at, i.updated_at,
		
		ili.line_item_id, ili.description, ili.quantity, ili.unit_price, ili.amount,
		ili.related_job_type, ili.related_job_id, ili.created_at AS item_created_at, ili.updated_at AS item_updated_at
	FROM invoices AS i
	JOIN clients AS c ON i.client_id = c.client_id
	LEFT JOIN staffs AS s ON i.assigned_staff_id = s.staff_id
	LEFT JOIN invoice_line_items AS ili ON i.invoice_id = ili.invoice_id`

	args := []interface{}{}
	paramCounter := 1

	if !isAdmin && staffIDFilter != "" { // Hanya filter jika bukan admin
		query += fmt.Sprintf(" WHERE i.assigned_staff_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
		paramCounter++
	}

	query += " ORDER BY i.invoice_date DESC, i.invoice_number DESC, ili.line_item_id ASC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all invoices: %w", err)
	}
	defer rows.Close()

	invoicesMap := make(map[string]*models.Invoice)
	var invoicesList []models.Invoice

	for rows.Next() {
		var (
			invoiceID         string
			invoiceNumber     string
			clientID          string
			clientName        string
			npwpClient        string
			assignedStaffID   sql.NullString
			assignedStaffName sql.NullString
			invoiceDate       time.Time
			dueDate           time.Time
			totalAmount       float64
			status            string
			notes             sql.NullString
			invoiceCreatedAt  time.Time
			invoiceUpdatedAt  time.Time

			lineItemID       sql.NullString
			description      sql.NullString
			quantity         sql.NullFloat64
			unitPrice        sql.NullFloat64
			amount           sql.NullFloat64
			relatedJobType   sql.NullString
			relatedJobID     sql.NullString
			itemCreatedAt    sql.NullTime
			itemUpdatedAt    sql.NullTime
		)

		err := rows.Scan(
			&invoiceID, &invoiceNumber, &clientID, &clientName, &npwpClient,
			&assignedStaffID, &assignedStaffName,
			&invoiceDate, &dueDate, &totalAmount, &status, &notes,
			&invoiceCreatedAt, &invoiceUpdatedAt,
			&lineItemID, &description, &quantity, &unitPrice, &amount,
			&relatedJobType, &relatedJobID, &itemCreatedAt, &itemUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice row: %w", err)
		}

		invoice, ok := invoicesMap[invoiceID]
		if !ok {
			invoice = &models.Invoice{
				InvoiceID:     invoiceID,
				InvoiceNumber: invoiceNumber,
				ClientID:      clientID,
				ClientName:    clientName,
				NpwpClient:    npwpClient,
				InvoiceDate:   models.CustomDate{Time: invoiceDate},
        DueDate:       models.CustomDate{Time: dueDate},
				TotalAmount:   totalAmount,
				Status:        status,
				CreatedAt:     invoiceCreatedAt,
				UpdatedAt:     invoiceUpdatedAt,
				LineItems:     []models.InvoiceLineItem{}, // Inisialisasi slice
			}
			if assignedStaffID.Valid {
				invoice.AssignedStaffID = &assignedStaffID.String
			} else {
				invoice.AssignedStaffID = nil
			}
			if assignedStaffName.Valid {
				invoice.AssignedStaffName = &assignedStaffName.String
			} else {
				invoice.AssignedStaffName = nil
			}
			if notes.Valid {
				invoice.Notes = &notes.String
			} else {
				invoice.Notes = nil
			}
			invoicesMap[invoiceID] = invoice
			invoicesList = append(invoicesList, *invoice)
		}

		// Add line item if exists
		if lineItemID.Valid {
			item := models.InvoiceLineItem{
				LineItemID:    lineItemID.String,
				InvoiceID:     invoiceID,
				Description:   description.String,
				Quantity:      quantity.Float64,
				UnitPrice:     unitPrice.Float64,
				Amount:        amount.Float64,
				CreatedAt:     itemCreatedAt.Time,
				UpdatedAt:     itemUpdatedAt.Time,
			}
			if relatedJobType.Valid {
				item.RelatedJobType = &relatedJobType.String
			} else {
				item.RelatedJobType = nil
			}
			if relatedJobID.Valid {
				item.RelatedJobID = &relatedJobID.String
			} else {
				item.RelatedJobID = nil
			}
			invoice.LineItems = append(invoice.LineItems, item)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for invoices: %w", err)
	}

	return invoicesList, nil
}

// GetInvoiceByID fetches a single invoice with its associated client, staff, and line items.
func (r *invoiceRepository) GetInvoiceByID(id string, staffIDFilter string, isAdmin bool) (*models.Invoice, error) {
	query := `
	SELECT
		i.invoice_id, i.invoice_number, i.client_id, c.client_name, c.npwp_client,
		i.assigned_staff_id, s.nama AS assigned_staff_name,
		i.invoice_date, i.due_date, i.total_amount, i.status, i.notes,
		i.created_at, i.updated_at,
		
		ili.line_item_id, ili.description, ili.quantity, ili.unit_price, ili.amount,
		ili.related_job_type, ili.related_job_id, ili.created_at AS item_created_at, ili.updated_at AS item_updated_at
	FROM invoices AS i
	JOIN clients AS c ON i.client_id = c.client_id
	LEFT JOIN staffs AS s ON i.assigned_staff_id = s.staff_id
	LEFT JOIN invoice_line_items AS ili ON i.invoice_id = ili.invoice_id
	WHERE i.invoice_id = $1`

	args := []interface{}{id}
	paramCounter := 2

	if !isAdmin && staffIDFilter != "" { // Hanya filter jika bukan admin
		query += fmt.Sprintf(" AND i.assigned_staff_id = $%d", paramCounter)
		args = append(args, staffIDFilter)
	}

	query += " ORDER BY ili.line_item_id ASC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by ID: %w", err)
	}
	defer rows.Close()

	var invoice *models.Invoice
	for rows.Next() {
		var (
			invoiceID         string
			invoiceNumber     string
			clientID          string
			clientName        string
			npwpClient        string
			assignedStaffID   sql.NullString
			assignedStaffName sql.NullString
			invoiceDate       time.Time
			dueDate           time.Time
			totalAmount       float64
			status            string
			notes             sql.NullString
			invoiceCreatedAt  time.Time
			invoiceUpdatedAt  time.Time

			lineItemID       sql.NullString
			description      sql.NullString
			quantity         sql.NullFloat64
			unitPrice        sql.NullFloat64
			amount           sql.NullFloat64
			relatedJobType   sql.NullString
			relatedJobID     sql.NullString
			itemCreatedAt    sql.NullTime
			itemUpdatedAt    sql.NullTime
		)

		err := rows.Scan(
			&invoiceID, &invoiceNumber, &clientID, &clientName, &npwpClient,
			&assignedStaffID, &assignedStaffName,
			&invoiceDate, &dueDate, &totalAmount, &status, &notes,
			&invoiceCreatedAt, &invoiceUpdatedAt,
			&lineItemID, &description, &quantity, &unitPrice, &amount,
			&relatedJobType, &relatedJobID, &itemCreatedAt, &itemUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice row by ID: %w", err)
		}

		if invoice == nil {
			invoice = &models.Invoice{
				InvoiceID:     invoiceID,
				InvoiceNumber: invoiceNumber,
				ClientID:      clientID,
				ClientName:    clientName,
				NpwpClient:    npwpClient,
				  InvoiceDate:   models.CustomDate{Time: invoiceDate}, // <-- PERBAIKAN DI SINI
        DueDate:       models.CustomDate{Time: dueDate},
				TotalAmount:   totalAmount,
				Status:        status,
				CreatedAt:     invoiceCreatedAt,
				UpdatedAt:     invoiceUpdatedAt,
				LineItems:     []models.InvoiceLineItem{}, // Inisialisasi slice
			}
			if assignedStaffID.Valid {
				invoice.AssignedStaffID = &assignedStaffID.String
			} else {
				invoice.AssignedStaffID = nil
			}
			if assignedStaffName.Valid {
				invoice.AssignedStaffName = &assignedStaffName.String
			} else {
				invoice.AssignedStaffName = nil
			}
			if notes.Valid {
				invoice.Notes = &notes.String
			} else {
				invoice.Notes = nil
			}
		}

		// Add line item if exists
		if lineItemID.Valid {
			item := models.InvoiceLineItem{
				LineItemID:    lineItemID.String,
				InvoiceID:     invoiceID,
				Description:   description.String,
				Quantity:      quantity.Float64,
				UnitPrice:     unitPrice.Float64,
				Amount:        amount.Float64,
				CreatedAt:     itemCreatedAt.Time,
				UpdatedAt:     itemUpdatedAt.Time,
			}
			if relatedJobType.Valid {
				item.RelatedJobType = &relatedJobType.String
			} else {
				item.RelatedJobType = nil
			}
			if relatedJobID.Valid {
				item.RelatedJobID = &relatedJobID.String
			} else {
				item.RelatedJobID = nil
			}
			invoice.LineItems = append(invoice.LineItems, item)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for invoice by ID: %w", err)
	}

	if invoice == nil {
		return nil, sql.ErrNoRows
	}

	return invoice, nil
}

// UpdateInvoice updates an existing invoice header info. Total amount is re-calculated.
func (r *invoiceRepository) UpdateInvoice(invoice *models.Invoice) error {
	// Re-calculate Total Amount based on current line items (assumes line items are managed externally or passed fully)
	// For a PATCH on header, you'd usually only update header fields.
	// If total_amount needs to be dynamic, it would be calculated in the handler based on fetched line items.
	// For simplicity, we'll assume line items are part of the Invoice struct if updating (though better to separate).
	// Or, more accurately, we would fetch existing line items and recalculate.
	// For now, let's update header fields explicitly.
	// We will not recalculate total_amount here, as it implies changing line items in this function.
	// Total amount should be updated when line items are added/removed/modified.

	var assignedStaffID sql.NullString
	if invoice.AssignedStaffID != nil && *invoice.AssignedStaffID != "" {
		assignedStaffID = sql.NullString{String: *invoice.AssignedStaffID, Valid: true}
	} else {
		assignedStaffID = sql.NullString{Valid: false}
	}

	var notes sql.NullString
	if invoice.Notes != nil && *invoice.Notes != "" {
		notes = sql.NullString{String: *invoice.Notes, Valid: true}
	} else {
		notes = sql.NullString{Valid: false}
	}

	query := `UPDATE invoices SET
		client_id = $1, assigned_staff_id = $2, invoice_date = $3, due_date = $4,
		total_amount = $5, status = $6, notes = $7, updated_at = $8
	WHERE invoice_id = $9`

	invoice.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		invoice.ClientID, assignedStaffID, invoice.InvoiceDate, invoice.DueDate,
		invoice.TotalAmount, invoice.Status, notes,
		invoice.UpdatedAt, invoice.InvoiceID,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

// DeleteInvoice deletes an invoice by its ID. Line items are deleted via CASCADE.
func (r *invoiceRepository) DeleteInvoice(id string) error {
	query := `DELETE FROM invoices WHERE invoice_id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

// CreateInvoiceLineItem inserts a new line item for an existing invoice.
func (r *invoiceRepository) CreateInvoiceLineItem(item *models.InvoiceLineItem) error {
	// Re-calculate amount for line item
	item.Amount = item.Quantity * item.UnitPrice

	// Handle nullable related_job fields
	var relatedJobType sql.NullString
	if item.RelatedJobType != nil && *item.RelatedJobType != "" {
		relatedJobType = sql.NullString{String: *item.RelatedJobType, Valid: true}
	} else {
		relatedJobType = sql.NullString{Valid: false}
	}

	var relatedJobID sql.NullString
	if item.RelatedJobID != nil && *item.RelatedJobID != "" {
		relatedJobID = sql.NullString{String: *item.RelatedJobID, Valid: true}
	} else {
		relatedJobID = sql.NullString{Valid: false}
	}

	query := `INSERT INTO invoice_line_items (
		invoice_id, description, quantity, unit_price, amount, related_job_type, related_job_id, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING line_item_id, created_at, updated_at`

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()

	err := r.db.QueryRow(query,
		item.InvoiceID, item.Description, item.Quantity, item.UnitPrice, item.Amount,
		relatedJobType, relatedJobID,
		item.CreatedAt, item.UpdatedAt,
	).Scan(&item.LineItemID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invoice line item: %w", err)
	}

	// Important: After creating a line item, the total_amount in the parent invoice needs to be updated.
	// This should be done in the handler or service layer.
	return nil
}

// UpdateInvoiceLineItem updates an existing line item.
func (r *invoiceRepository) UpdateInvoiceLineItem(item *models.InvoiceLineItem) error {
	// Re-calculate amount for line item if quantity or unit_price are updated (handled in handler)
	// For PATCH, quantity/unit_price might be nil. Amount should be calculated in handler if they are provided.
	
	// Handle nullable related_job fields
	var relatedJobType sql.NullString
	if item.RelatedJobType != nil && *item.RelatedJobType != "" {
		relatedJobType = sql.NullString{String: *item.RelatedJobType, Valid: true}
	} else {
		relatedJobType = sql.NullString{Valid: false}
	}

	var relatedJobID sql.NullString
	if item.RelatedJobID != nil && *item.RelatedJobID != "" {
		relatedJobID = sql.NullString{String: *item.RelatedJobID, Valid: true}
	} else {
		relatedJobID = sql.NullString{Valid: false}
	}

	query := `UPDATE invoice_line_items SET
		description = $1, quantity = $2, unit_price = $3, amount = $4,
		related_job_type = $5, related_job_id = $6, updated_at = $7
	WHERE line_item_id = $8`

	item.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		item.Description, item.Quantity, item.UnitPrice, item.Amount,
		relatedJobType, relatedJobID,
		item.UpdatedAt, item.LineItemID,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice line item: %w", err)
	}
	// Important: After updating a line item, the total_amount in the parent invoice might need to be updated.
	return nil
}

// DeleteInvoiceLineItem deletes a line item by its ID.
func (r *invoiceRepository) DeleteInvoiceLineItem(itemID string) error {
	query := `DELETE FROM invoice_line_items WHERE line_item_id = $1`
	_, err := r.db.Exec(query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice line item: %w", err)
	}
	// Important: After deleting a line item, the total_amount in the parent invoice needs to be updated.
	return nil
}