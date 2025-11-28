package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// album represents data about a record album.
type album struct {
	gorm.Model
	Title  string
	Artist string
	Price  float64
}

// response to user to hide some db fields
type AlbumResponse struct {
	ID     uint    `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// global db variable to use outside of main
var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("albums.db"), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	db.AutoMigrate(&album{})

	router := gin.Default()

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbum)
	router.POST("/albums", createUpdateAlbum)
	router.PATCH("/albums/:id", createUpdateAlbum)
	router.DELETE("/albums/:id", deleteAlbum)

	router.Run()

}

func getAlbums(ctx *gin.Context) {
	var albums []album
	var albumsRes []AlbumResponse

	if result := db.Find(&albums); result.Error != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "No albums saved"})
		return
	}

	for _, a := range albums {
		albumsRes = append(albumsRes, getAlbumRes(a))
	}

	ctx.IndentedJSON(http.StatusOK, albumsRes)
}

func getAlbum(ctx *gin.Context) {

	var album album
	id := getId(ctx)
	if id == -1 {
		return
	}
	if result := db.First(&album, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Could not find album"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, getAlbumRes(album))
}

func createUpdateAlbum(ctx *gin.Context) {
	var newAlbum album
	var album album

	if err := ctx.BindJSON(&newAlbum); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	switch ctx.Request.Method {
	case http.MethodPost:

		db.Create(&newAlbum)

		ctx.IndentedJSON(http.StatusCreated, getAlbumRes(newAlbum))

	case http.MethodPatch:
		id := getId(ctx)
		if id == -1 {
			return
		}
		if result := db.First(&album, id); result.Error != nil {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Could not find album"})
			return
		}
		db.Model(&album).Select("Title", "Artist", "Price").Updates(newAlbum)

		ctx.IndentedJSON(http.StatusOK, getAlbumRes(album))
	}

}

func deleteAlbum(ctx *gin.Context) {
	var album album
	id := getId(ctx)
	if id == -1 {
		return
	}

	if result := db.First(&album, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "Could not find album"})
		return
	}

	// soft delete - use Unscoped() to delete permanently
	db.Delete(&album)
	//db.Unscoped().Delete(&album)
	//db.Delete(&album{}, id)

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"message": "Album deleted successfully",
		"id":      album.ID,
	})
}

func getAlbumRes(a album) AlbumResponse {
	return AlbumResponse{ID: a.ID, Title: a.Title, Artist: a.Artist, Price: a.Price}
}

func getId(ctx *gin.Context) int {
	strId, strErr := ctx.Params.Get("id")
	id, err := strconv.Atoi(strId)
	if !strErr || err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid param passed"})
		return -1
	}
	return id
}
