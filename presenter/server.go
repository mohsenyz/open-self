package presenter

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"openSelf/data"
	"path"
	"strings"
)

func Index(c *gin.Context) {
	c.FileFromFS("favicon.png", data.Assets)
}

type CustomFS struct {
	fs http.FileSystem
}

func (fs CustomFS) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join("/", p)
		file, err := fs.Open(name)
		if err != nil {
			return false
		}
		stats, err := file.Stat()
		if err != nil {
			return false
		}
		if stats.IsDir() {
			return false
		}
		return true
	}
	return false
}

func (fs CustomFS) Open(addr string) (file http.File, err error) {
	return fs.fs.Open(addr)
}

func StartServer()  {
	server := gin.Default()
	server.Use(static.Serve("/", CustomFS{fs: data.Assets}))
	server.Run()
}