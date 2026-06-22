package server

import "github.com/gin-gonic/gin"

const UserIdKey = "user_id"

func GetUserId(c *gin.Context) (string, bool) {
	v, _ := c.Get(UserIdKey)
	s, ok := v.(string)
	return s, ok
}

func SetUserId(c *gin.Context, uid string) {
	c.Set(UserIdKey, uid)
}
