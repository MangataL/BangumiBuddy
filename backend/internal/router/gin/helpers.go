package gin

import (
	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

func writeError(c *gin.Context, err error) {
	code, msg := errs.ParseError(err)
	c.JSON(code, gin.H{"error": msg})
}
