package websocket

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 將一般的 HTTP 連線升級為 WebSocket 連線
var upgrader = websocket.Upgrader{
	// 任何請求來源允許，僅能在開發測試過程中使用
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Message WebSocket 訊息結構
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// Manager WebSocket 連線管理器
type Manager struct {
	clients   map[*websocket.Conn]bool // 連線的客戶端
	broadcast chan Message             // 廣播訊息的通道
	messages  []Message               // 訊息歷史
	mu        sync.Mutex              // 保護共享資源的互斥鎖
}

// NewManager 創建新的 WebSocket 管理器
func NewManager() *Manager {
	return &Manager{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message),
		messages:  make([]Message, 0),
	}
}

// Run 啟動 WebSocket 管理器（在 goroutine 中運行）
func (m *Manager) Run() {
	for {
		msg := <-m.broadcast
		
		// 保存訊息到歷史記錄
		m.mu.Lock()
		m.messages = append(m.messages, msg)
		m.mu.Unlock()
		
		// 廣播訊息給所有客戶端
		m.mu.Lock()
		for client := range m.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(m.clients, client)
			}
		}
		m.mu.Unlock()
	}
}

// HandleConnection 處理新的 WebSocket 連線
func (m *Manager) HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	
	// 註冊新客戶端
	m.mu.Lock()
	m.clients[conn] = true
	m.mu.Unlock()
	
	// 發送歷史訊息給新客戶端
	m.mu.Lock()
	for _, msg := range m.messages {
		conn.WriteJSON(msg)
	}
	m.mu.Unlock()
	
	// 處理客戶端訊息
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			// 客戶端斷線，清理資源
			m.mu.Lock()
			delete(m.clients, conn)
			m.mu.Unlock()
			conn.Close()
			break
		}
		
		// 將訊息發送到廣播通道
		m.broadcast <- msg
	}
}

// GetMessages 獲取所有訊息歷史
func (m *Manager) GetMessages() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 返回訊息副本
	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// AddMessage 添加新訊息
func (m *Manager) AddMessage(msg Message) {
	m.mu.Lock()
	m.messages = append(m.messages, msg)
	m.mu.Unlock()
}

// GetClientCount 獲取當前連線數
func (m *Manager) GetClientCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.clients)
}
