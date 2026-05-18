package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/prodonik/bank_app/internal/infrastructure/auth"
	"github.com/prodonik/bank_app/internal/interfaces/api/handler"
	"github.com/prodonik/bank_app/internal/interfaces/api/middleware"
)

func NewRouter(userHandler *handler.UserHandler, cityHandler *handler.CityHandler, innHandler *handler.InnHandler, ifutCodeHandler *handler.IfutCodeHandler, entrepreneurHandler *handler.EntrepreneurHandler, jwtService *auth.JWTService) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/login", userHandler.Login)
			users.POST("/refresh", userHandler.Refresh)

			authenticated := users.Group("")
			authenticated.Use(middleware.AuthMiddleware(jwtService))
			{
				authenticated.POST("/register", userHandler.Register)
				authenticated.POST("/logout", userHandler.Logout)
				authenticated.GET("/me", userHandler.GetMe)
				authenticated.GET("", userHandler.GetAll)
				authenticated.GET("/:id", userHandler.GetByID)
				authenticated.PUT("/:id", userHandler.Update)
				authenticated.DELETE("/:id", userHandler.Delete)
			}
		}

		cities := v1.Group("/cities")
		cities.Use(middleware.AuthMiddleware(jwtService))
		{
			cities.POST("", cityHandler.Create)
			cities.GET("", cityHandler.GetAll)
			cities.GET("/:id", cityHandler.GetByID)
			cities.PUT("/:id", cityHandler.Update)
			cities.DELETE("/:id", cityHandler.Delete)
		}

		inns := v1.Group("/inns")
		inns.Use(middleware.AuthMiddleware(jwtService))
		{
			inns.GET("", innHandler.GetAll)
		}

		ifutCodes := v1.Group("/ifut-codes")
		ifutCodes.Use(middleware.AuthMiddleware(jwtService))
		{
			ifutCodes.GET("", ifutCodeHandler.GetAll)
		}

		entrepreneurs := v1.Group("/entrepreneurs")
		{
			entrepreneurs.PUT("/birdarcha-token", entrepreneurHandler.UpdateBirdarchaToken)

			authenticated := entrepreneurs.Group("")
			authenticated.Use(middleware.AuthMiddleware(jwtService))
			{
				authenticated.POST("", entrepreneurHandler.Create)
				authenticated.GET("", entrepreneurHandler.GetAll)
				authenticated.GET("/sqb-failed", entrepreneurHandler.GetSqbFailed)
				authenticated.POST("/sqb-retry", entrepreneurHandler.RetrySqbFailed)
				authenticated.GET("/:id", entrepreneurHandler.GetByID)
				authenticated.PUT("/:id", entrepreneurHandler.Update)
				authenticated.DELETE("/:id", entrepreneurHandler.Delete)
			}
		}
	}

	return r
}
