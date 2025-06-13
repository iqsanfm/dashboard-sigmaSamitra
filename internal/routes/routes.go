package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/config"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/handlers"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/middlewares"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/repositories"
)

// SetupRouter sets up all application routes
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Initialize repositories
	clientRepo := repositories.NewClientRepository(db)
	monthlyJobRepo := repositories.NewMonthlyJobRepository(db)
	staffRepo := repositories.NewStaffRepository(db)
	annualJobRepo := repositories.NewAnnualJobRepository(db)
	sp2dkJobRepo := repositories.NewSp2dkJobRepository(db)
	pemeriksaanJobRepo := repositories.NewPemeriksaanJobRepository(db)

	// Initialize handlers
	clientHandler := handlers.NewClientHandler(clientRepo, staffRepo, monthlyJobRepo, annualJobRepo, sp2dkJobRepo, pemeriksaanJobRepo) 
	monthlyJobHandler := handlers.NewMonthlyJobHandler(monthlyJobRepo, clientRepo, staffRepo)
	staffHandler := handlers.NewStaffHandler(staffRepo)
	authHandler := handlers.NewAuthHandler(staffRepo, cfg.JWTSecretKey)
	annualJobHandler := handlers.NewAnnualJobHandler(annualJobRepo, clientRepo, staffRepo)
	sp2dkJobHandler := handlers.NewSp2dkJobHandler(sp2dkJobRepo, clientRepo, staffRepo)
	pemeriksaanJobHandler := handlers.NewPemeriksaanJobHandler(pemeriksaanJobRepo, clientRepo, staffRepo) 

	// Initialize Auth Middleware
	authMiddleware := middlewares.NewAuthMiddleware(cfg)

	// API V1 Group
	v1 := r.Group("/api/v1")

	
	{
		// Auth routes (NOT PROTECTED)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.Login)
		}

		// Protected Routes Group
		protected := v1.Group("/")
		// --- JANGAN LUPA AKTIFKAN KEMBALI AUTENTIKASI SETELAH SELESAI DEVELOPMENT ---
		protected.Use(authMiddleware.AuthRequired())
		// --- AKHIR PENGINGAT ---

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

			r.Static("/uploads", "./uploads")

			// Monthly Job routes
			monthlyJobRoutes := protected.Group("/monthly-jobs")
			{
				monthlyJobRoutes.POST("/", monthlyJobHandler.CreateMonthlyJob)
				monthlyJobRoutes.GET("/", monthlyJobHandler.GetAllMonthlyJobs)
				monthlyJobRoutes.GET("/:id", monthlyJobHandler.GetMonthlyJobByID)
				monthlyJobRoutes.PATCH("/:id", monthlyJobHandler.UpdateMonthlyJob)
				

				// Nested Tax Report routes
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

				// Nested SPT Tahunan Reports
				sptRoutes := annualJobRoutes.Group("/:id/spt-reports")
				{
					sptRoutes.POST("/", annualJobHandler.CreateAnnualTaxReport)
					sptRoutes.PATCH("/:report_id", annualJobHandler.UpdateAnnualTaxReport)
					sptRoutes.DELETE("/:report_id", annualJobHandler.DeleteAnnualTaxReport)
				}

				// Nested Dividend Reports
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

			// SP2DK Job routes <-- RUTE BARU UNTUK SP2DK
			sp2dkJobRoutes := protected.Group("/sp2dk-jobs")
			{
				sp2dkJobRoutes.POST("/", sp2dkJobHandler.CreateSp2dkJob)
				sp2dkJobRoutes.GET("/", sp2dkJobHandler.GetAllSp2dkJobs)
				sp2dkJobRoutes.GET("/:id", sp2dkJobHandler.GetSp2dkJobByID)
				sp2dkJobRoutes.PATCH("/:id", sp2dkJobHandler.UpdateSp2dkJob)
				sp2dkJobRoutes.DELETE("/:id", sp2dkJobHandler.DeleteSp2dkJob)
			}

			pemeriksaanJobRoutes := protected.Group("/pemeriksaan-jobs")
			{
				pemeriksaanJobRoutes.POST("/", pemeriksaanJobHandler.CreatePemeriksaanJob)
				pemeriksaanJobRoutes.GET("/", pemeriksaanJobHandler.GetAllPemeriksaanJobs)
				pemeriksaanJobRoutes.GET("/:id", pemeriksaanJobHandler.GetPemeriksaanJobByID)
				pemeriksaanJobRoutes.PATCH("/:id", pemeriksaanJobHandler.UpdatePemeriksaanJob)
				pemeriksaanJobRoutes.DELETE("/:id", pemeriksaanJobHandler.DeletePemeriksaanJob)
			}
		}
	}

	return r
}