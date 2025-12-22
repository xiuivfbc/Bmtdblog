package router

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

func setTemplate(engine *gin.Engine) {
	funcMap := template.FuncMap{
		"dateFormat": common.DateFormat,
		"substring":  common.Substring,
		"isOdd":      common.IsOdd,
		"isEven":     common.IsEven,
		"truncate":   common.Truncate,
		"length":     common.Len,
		"add":        common.Add,
		"sub":        common.Sub,
		"minus":      common.Minus,
		"multiply":   common.Multiply,
		"seq":        common.Seq,
		"listtag":    common.ListTag,
	}
	engine.SetFuncMap(funcMap)
	engine.LoadHTMLGlob(common.GetCurrentDirectory() + "/front/views/**/*.html")
}
