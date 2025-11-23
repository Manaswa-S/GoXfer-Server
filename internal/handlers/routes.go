package handler

import (
	"fmt"
	"goxfer/server/internal/consts/errs"
	"goxfer/server/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(srv *service.Service) *Handler {
	return &Handler{
		service: srv,
	}
}

func (h *Handler) RegisterRoutes(publicGrp, privateGrp *gin.RouterGroup) error {

	publicGrp.GET("/bucket/open/config", h.getConfigs)

	// initiate a bucket creation, opaque step 1, registration request
	publicGrp.POST("/bucket/create/s1", h.createBucketS1)
	// complete the bucket creation, opaque step 2, registration record
	publicGrp.POST("/bucket/create/s2", h.createBucketS2)

	// initiate a bucket opening, opaque step 1, login init
	publicGrp.POST("/bucket/open/s1", h.openBucketS1)
	// complete a bucket opening, opaque step 2, login finish
	publicGrp.POST("/bucket/open/s2", h.openBucketS2)

	// initiate a new file upload
	privateGrp.POST("/file/upload/init", h.initUpload)
	// upload file parts
	privateGrp.POST("/file/upload/part", h.uploadPart)
	// complete the file upload
	privateGrp.POST("/file/upload/complete", h.completeUpload)

	// get the complete file list for a bucket key
	privateGrp.GET("/file/list", h.getFilesList)

	privateGrp.GET("/file/download/init", h.downloadInit)
	// get/download the file (still encrypted)
	privateGrp.GET("/file/download/data", h.downloadData)
	// get the metadata to decrypt file
	privateGrp.GET("/file/download/meta", h.downloadMeta)
	// get the digest and all to verify file
	privateGrp.GET("/file/download/digest", h.downloadDigest)

	privateGrp.DELETE("/file/delete", h.deleteFile)

	if err := h.registerAuxiliaryRoutes(publicGrp); err != nil {
		return err
	}
	return nil
}

func (h *Handler) registerAuxiliaryRoutes(publicGrp *gin.RouterGroup) error {

	publicGrp.POST("/test/upload", h.testUpload)
	publicGrp.GET("/test/download", h.testDownload)

	return nil
}

func (h *Handler) handleErrf(ctx *gin.Context, errf *errs.Errorf) bool {
	if errf == nil {
		return false
	}

	if errf.ReturnRaw {
		ctx.JSON(http.StatusBadRequest, errf)
	} else {
		fmt.Println(errf)
		ctx.Status(http.StatusInternalServerError)
	}
	return true
}

// TODO: expose routes.
// func RoutesHandler(e *echo.Echo) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         return c.JSON(http.StatusOK, e.Routes())
//     }
// }
