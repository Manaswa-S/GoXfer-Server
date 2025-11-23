package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// These handlers do not have service layer methods.
// Everything happens in here, and therefore are
// usually not linked to the database.

func (h *Handler) testUpload(ctx *gin.Context) {
	r := bufio.NewReader(ctx.Request.Body)
	lenStr, err := r.ReadString('\n')
	if err != nil {
		fmt.Println("failed to read length:", err)
		return
	}

	upLen, err := strconv.ParseInt(strings.TrimSuffix(lenStr, "\n"), 10, 64)
	if err != nil {
		fmt.Println("invalid length:", err)
		return
	}

	chunkSize := int64(1024 * 1024)
	totalUp := int64(0)

	for totalUp < upLen {
		n, err := io.CopyN(io.Discard, r, chunkSize)
		totalUp += n

		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			return
		}
	}
}

func (h *Handler) testDownload(ctx *gin.Context) {
	downLen := 4 * 1024 * 1024
	downData := bytes.Repeat([]byte{0xAA}, downLen)

	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-store")
	ctx.Writer.Header().Add("Content-Length", fmt.Sprintf("%d", downLen))
	ctx.Writer.Header().Add("Start-Time", fmt.Sprintf("%d", time.Now().UnixMilli()))

	n, err := ctx.Writer.Write(downData)
	if err != nil || n != downLen {
		return
	}
}
