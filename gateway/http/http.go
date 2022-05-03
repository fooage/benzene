package gateway

import (
	"log"
	"net"
	"net/http"

	"github.com/fooage/benzene/cache"
	"github.com/fooage/benzene/service"
	"github.com/gin-gonic/gin"
)

// This is the abstraction of the HTTP interface layer to the cache data
// structure. Only used for binding of parameters!

type Item struct {
	Key   string `json:"key"`
	Value Value  `json:"value"`
}

type Value string

func (v Value) Raw() *[]byte {
	raw := []byte(v)
	return &raw
}

func ServeWithHTTP(addr net.Addr) {
	r := gin.Default()
	r.POST("/get", redirect, get)
	r.POST("/set", redirect, set)
	if err := r.Run(addr.String()); err != nil {
		log.Fatalf("Http server run error: %v\n", err)
		return
	}
}

func redirect(ctx *gin.Context) {
	var item Item
	err := ctx.Bind(&item)
	if err != nil {
		log.Panicf("Redirect middleware bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "request param error",
			"data": nil,
		})
		ctx.Abort()
		return
	}
	addr, err := service.Guider.PickPeer(item.Key, service.Info)
	if err != nil {
		log.Printf("Redirect pick peer error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code": 2,
			"msg":  "cluster error",
			"data": nil,
		})
		ctx.Abort()
		return
	}
	if addr.String() == service.Info.Addr.String() {
		// set the params to context
		ctx.Set("electron", item)
		ctx.Next()
	} else {
		carbon := addr.String() + ctx.FullPath()
		log.Printf("Redirect to %v carbon", addr)
		ctx.Redirect(http.StatusTemporaryRedirect, carbon)
	}
}

func get(ctx *gin.Context) {
	params, exist := ctx.Get("electron")
	if !exist {
		ctx.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "request param error",
			"data": nil,
		})
		return
	}
	item := params.(Item)
	res, ok := cache.Carbon.GET(item.Key)
	if ok {
		// hit cache
		ctx.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "hit cache",
			"data": res,
		})
	} else {
		// not hit cache
		ctx.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "not hit cache",
			"data": nil,
		})
	}
}

func set(ctx *gin.Context) {
	params, exist := ctx.Get("electron")
	if !exist {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "request param error",
			"data": nil,
		})
		return
	}
	item := params.(Item)
	cache.Carbon.SET(item.Key, item.Value)
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "set cache",
		"data": nil,
	})
}
