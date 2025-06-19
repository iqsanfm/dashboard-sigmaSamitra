package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/config"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/handlers"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/middlewares"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/services"
)

// SetupRouter sets up all application routes
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// 1. Initialize Repositories
clientRepo := repositories.NewClientRepository(db)
	monthlyJobRepo := repositories.NewMonthlyJobRepository(db)
	staffRepo := repositories.NewStaffRepository(db)
	annualJobRepo := repositories.NewAnnualJobRepository(db)
	sp2dkJobRepo := repositories.NewSp2dkJobRepository(db)
	pemeriksaanJobRepo := repositories.NewPemeriksaanJobRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)

		invoiceService := services.NewInvoiceService(
		invoiceRepo,
		monthlyJobRepo,
		annualJobRepo,
		sp2dkJobRepo,
		pemeriksaanJobRepo,
	)

	// 2. Initialize Handlers
	clientHandler := handlers.NewClientHandler(clientRepo, staffRepo, monthlyJobRepo, annualJobRepo, sp2dkJobRepo, pemeriksaanJobRepo)
	staffHandler := handlers.NewStaffHandler(staffRepo)
	authHandler := handlers.NewAuthHandler(staffRepo, cfg.JWTSecretKey)
	monthlyJobHandler := handlers.NewMonthlyJobHandler(monthlyJobRepo, clientRepo, staffRepo, invoiceService)
	annualJobHandler := handlers.NewAnnualJobHandler(annualJobRepo, clientRepo, staffRepo, invoiceService)
	sp2dkJobHandler := handlers.NewSp2dkJobHandler(sp2dkJobRepo, clientRepo, staffRepo, invoiceService)
	pemeriksaanJobHandler := handlers.NewPemeriksaanJobHandler(pemeriksaanJobRepo, clientRepo, staffRepo, invoiceService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo)

	// 3. Initialize Middleware
	authMiddleware := middlewares.NewAuthMiddleware(cfg)

	// 4. Setup Route Groups
	v1 := r.Group("/api/v1")
	{
		// Auth routes (Public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.Login)
		}

		// Protected Routes Group
		protected := v1.Group("/")
		protected.Use(authMiddleware.AuthRequired())
		{
			// Client routes
			clientRoutes := protected.Group("/clients")
			{
				clientRoutes.POST("/", clientHandler.CreateClient)
				clientRoutes.GET("/", clientHandler.GetAllClients)
				clientRoutes.GET("/:id", clientHandler.GetClientByID)
				clientRoutes.GET("/:id/all-jobs", clientHandler.GetClientDashboardJobs)
				clientRoutes.PATCH("/:id", clientHandler.UpdateClient)
				clientRoutes.DELETE("/:id", clientHandler.DeleteClient)
			}

			// Static route for file uploads
			r.Static("/uploads", "./uploads")

			// Monthly Job routes
			monthlyJobRoutes := protected.Group("/monthly-jobs")
			{
				monthlyJobRoutes.POST("/", monthlyJobHandler.CreateMonthlyJob)
				monthlyJobRoutes.GET("/", monthlyJobHandler.GetAllMonthlyJobs)
				monthlyJobRoutes.GET("/:id", monthlyJobHandler.GetMonthlyJobByID)
				monthlyJobRoutes.PATCH("/:id", monthlyJobHandler.UpdateMonthlyJob)

				taxReportRoutes := monthlyJobRoutes.Group("/:id/tax-reports")
				{
					taxReportRoutes.POST("/", monthlyJobHandler.CreateMonthlyTaxReport)
					taxReportRoutes.PATCH("/:report_id", monthlyJobHandler.UpdateMonthlyTaxReport)
					taxReportRoutes.DELETE("/:report_id", monthlyJobHandler.DeleteMonthlyTaxReport)
				}
			}

			// Annual Job routes
			annualJobRoutes := protected.Group("/annual-jobs")
			{
				annualJobRoutes.POST("/", annualJobHandler.CreateAnnualJob)
				annualJobRoutes.GET("/", annualJobHandler.GetAllAnnualJobs)
				annualJobRoutes.GET("/:id", annualJobHandler.GetAnnualJobByID)
				annualJobRoutes.PATCH("/:id", annualJobHandler.UpdateAnnualJob)

				sptRoutes := annualJobRoutes.Group("/:id/spt-reports")
				{
					sptRoutes.POST("/", annualJobHandler.CreateAnnualTaxReport)
					sptRoutes.PATCH("/:report_id", annualJobHandler.UpdateAnnualTaxReport)
					sptRoutes.DELETE("/:report_id", annualJobHandler.DeleteAnnualTaxReport)
				}

				dividendRoutes := annualJobRoutes.Group("/:id/dividend-reports")
				{
					dividendRoutes.POST("/", annualJobHandler.CreateAnnualDividendReport)
					dividendRoutes.PATCH("/:report_id", annualJobHandler.UpdateAnnualDividendReport)
					dividendRoutes.DELETE("/:report_id", annualJobHandler.DeleteAnnualDividendReport)
				}
			}

			// Staff routes
			staffRoutes := protected.Group("/staffs")
			{
				staffRoutes.POST("/", staffHandler.CreateStaff)
				staffRoutes.GET("/", staffHandler.GetAllStaffs)
				staffRoutes.GET("/:id", staffHandler.GetStaffByID)
				staffRoutes.PATCH("/:id", staffHandler.UpdateStaff)
				staffRoutes.DELETE("/:id", staffHandler.DeleteStaff)
				staffRoutes.PATCH("/:id/password", staffHandler.ChangeStaffPassword)
			}

			// SP2DK Job routes
			sp2dkJobRoutes := protected.Group("/sp2dk-jobs")
			{
				sp2dkJobRoutes.POST("/", sp2dkJobHandler.CreateSp2dkJob)
				sp2dkJobRoutes.GET("/", sp2dkJobHandler.GetAllSp2dkJobs)
				sp2dkJobRoutes.GET("/:id", sp2dkJobHandler.GetSp2dkJobByID)
				sp2dkJobRoutes.PATCH("/:id", sp2dkJobHandler.UpdateSp2dkJob)
				sp2dkJobRoutes.DELETE("/:id", sp2dkJobHandler.DeleteSp2dkJob)
			}

			// Pemeriksaan Job routes
			pemeriksaanJobRoutes := protected.Group("/pemeriksaan-jobs")
			{
				pemeriksaanJobRoutes.POST("/", pemeriksaanJobHandler.CreatePemeriksaanJob)
				pemeriksaanJobRoutes.GET("/", pemeriksaanJobHandler.GetAllPemeriksaanJobs)
				pemeriksaanJobRoutes.GET("/:id", pemeriksaanJobHandler.GetPemeriksaanJobByID)
				pemeriksaanJobRoutes.PATCH("/:id", pemeriksaanJobHandler.UpdatePemeriksaanJob)
				pemeriksaanJobRoutes.DELETE("/:id", pemeriksaanJobHandler.DeletePemeriksaanJob)
			}

			// Invoice Routes
			invoiceRoutes := protected.Group("/invoices")
			{
				invoiceRoutes.POST("/", invoiceHandler.CreateInvoice)
				invoiceRoutes.GET("/", invoiceHandler.GetAllInvoices)
				invoiceRoutes.GET("/:id", invoiceHandler.GetInvoiceByID)
			}
		}
	}

	return r
}