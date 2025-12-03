package handler

import (
	"fmt"
	"goxfer/server/internal/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getWelcome(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func (h *Handler) createBucketS1(ctx *gin.Context) {
	req := new(dto.CreateBucketS1Req)
	if err := ctx.Bind(req); err != nil {
		return
	}

	resp, errf := h.service.CreateBucketS1(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) createBucketS2(ctx *gin.Context) {
	req := new(dto.CreateBucketS2Req)
	if err := ctx.Bind(req); err != nil {
		return
	}

	resp, errf := h.service.CreateBucketS2(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) getConfigs(ctx *gin.Context) {
	resp, errf := h.service.GetConfigs(ctx)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) openBucketS1(ctx *gin.Context) {
	req := new(dto.OpenBucketS1Req)
	if err := ctx.Bind(req); err != nil {
		return
	}

	resp, errf := h.service.OpenBucketS1(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) openBucketS2(ctx *gin.Context) {
	req := new(dto.OpenBucketS2Req)
	if err := ctx.Bind(req); err != nil {
		return
	}

	resp, errf := h.service.OpenBucketS2(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) getBucketData(ctx *gin.Context) {
	resp, errf := h.service.GetBucketData(ctx)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) initUpload(ctx *gin.Context) {
	req := new(dto.InitUploadReq)
	if err := ctx.Bind(req); err != nil {
		fmt.Println(err)
		return
	}

	resp, errf := h.service.InitUpload(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func (h *Handler) uploadPart(ctx *gin.Context) {
	uploadId := ctx.Query("upload_id")
	if uploadId == "" {
		return
	}
	chunkId := ctx.Query("chunk_id")
	if chunkId == "" {
		return
	}
	errf := h.service.UploadPart(ctx, uploadId, chunkId)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.Status(http.StatusAccepted)
}

func (h *Handler) completeUpload(ctx *gin.Context) {
	req := new(dto.CompleteUploadReq)
	if err := ctx.Bind(req); err != nil {
		fmt.Println(err)
		return
	}

	errf := h.service.CompleteUpload(ctx, req)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) getFilesList(ctx *gin.Context) {
	resp, errf := h.service.GetFilesList(ctx)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) downloadInit(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		return
	}

	resp, errf := h.service.DownloadInit(ctx, fileID)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) downloadData(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		return
	}

	errf := h.service.DownloadData(ctx, fileID)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) downloadMeta(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		return
	}

	resp, errf := h.service.DownloadMeta(ctx, fileID)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) downloadDigest(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		return
	}

	resp, errf := h.service.DownloadDigest(ctx, fileID)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) deleteFile(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		return
	}

	errf := h.service.DeleteFile(ctx, fileID)
	if h.handleErrf(ctx, errf) {
		return
	}

	ctx.Status(http.StatusOK)
}
