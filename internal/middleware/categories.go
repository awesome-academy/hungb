package middleware

import (
	"sun-booking-tours/internal/repository"

	"github.com/gin-gonic/gin"
)

const contextKeyNavCategories = "nav_categories"

func LoadCategories(catRepo repository.CategoryRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		cats, err := catRepo.FindAllParents(c.Request.Context())
		if err == nil {
			c.Set(contextKeyNavCategories, cats)
		}
		c.Next()
	}
}

func GetNavCategories(c *gin.Context) any {
	v, _ := c.Get(contextKeyNavCategories)
	return v
}
