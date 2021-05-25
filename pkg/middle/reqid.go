package middle

import (
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func ReqId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("reqId", uuid.NewUUID())
	}
}
