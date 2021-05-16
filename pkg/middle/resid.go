package middle

import (
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func ResId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("resId", uuid.NewUUID())
	}
}
