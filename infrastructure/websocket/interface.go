package websocket

import "github.com/gin-gonic/gin"

// WebSocketManager WebSocket 管理器介面
// 這個介面定義了 WebSocket 管理的核心功能
type WebSocketManager interface {
	// Run 啟動管理器（在 goroutine 中運行）
	Run()

	// HandleConnection 處理新的 WebSocket 連線
	HandleConnection(c *gin.Context)

	// GetMessages 獲取所有訊息歷史
	GetMessages() []Message

	// AddMessage 添加新訊息
	AddMessage(msg Message)

	// GetClientCount 獲取當前連線數
	GetClientCount() int
}

// 確保 Manager 實現了 WebSocketManager 介面
var _ WebSocketManager = (*Manager)(nil)
