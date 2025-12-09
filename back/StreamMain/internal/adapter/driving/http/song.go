package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/egors-prof/streaming/pkg"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID       int
	conn     *websocket.Conn
	write    chan string
	isClosed bool
}

var clients = make(map[int]*Client)

var clientCount = 0

func (s *Server) WebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.handleError(c, err)
	}
	newClient := &Client{ID: clientCount, conn: conn, write: make(chan string), isClosed: false}
	clients[clientCount] = newClient
	clientCount++
	fmt.Println("established websocket connection")
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			s.handleError(c, err)
		}
	}(conn)
	go func() {
		err := newClient.WaitForMessage()
		if err != nil {
			s.handleError(c, err)
		}
	}()

	select {}
}

type authStruct struct {
	DataType string `json:"data_type"`
	Token    string `json:"access_token"`
}

func countChar(str string, char string) int {
	count := 0
	for _, c := range str {
		if string(c) == char {
			count++
		}
	}
	return count
}

func (c *Client) WaitForMessage() error {
	fmt.Println("goroutine spawned")
	defer func() {
		c.disconnect()
	}()
	///
	_, message, err := c.conn.ReadMessage()
	fmt.Println(string(message))
	if err != nil {
		return err
	}
	if strings.Contains(string(message), "auth") && (countChar(string(message), ".")) == 2 {
		auth := authStruct{}
		if err := json.Unmarshal(message, &auth); err != nil {
			log.Println("unmarshal err:", err)
			return err
		}
		fmt.Println(auth)
		userID, username, isRefresh, userRole, err := pkg.ParseToken(auth.Token)

		if err != nil {
			return err
		}

		if userID == 0 {
			log.Println("parse err:", err)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte("authentication error"))
			if err != nil {
				return err
			}

		}
		bytes, err := json.Marshal(struct {
			UserID    int    `json:"user_id"`
			Username  string `json:"username"`
			Role      string `json:"role"`
			IsRefresh bool   `json:"isRefresh"`
		}{UserID: userID, Username: username, Role: string(userRole), IsRefresh: isRefresh})
		err = c.conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			return err
		}
		err = c.conn.WriteMessage(websocket.TextMessage, []byte("auth success"))
		if err != nil {
			return err
		}

	} else {
		return err
	}
	///
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		if len(message) > 5 {
			strMessage := string(message)
			log.Println(strMessage)
			log.Println(strMessage[5:])
			if strMessage[:5] == "play:" {
				c.StreamFile(strMessage[5:])
			} else {
				continue
			}

		}

	}
	return nil
}

func (c *Client) StreamFile(title string) {

	fmt.Println("got a signal")
	file, err := os.Open(fmt.Sprintf("music/%s.wav", title))
	if err != nil {
		log.Println("wav reading error ", err)
		err := c.conn.WriteMessage(websocket.TextMessage, []byte("error: "+err.Error()))
		if err != nil {
			return
		}
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fmt.Printf("our file is %d bytes", fileSize)
	err = c.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("file %d bytes", fileSize)))
	if err != nil {
		return
	}
	buffer := make([]byte, 16384) //16 kilobytes
	var totalSent int64 = 0
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			err := c.conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
			if err != nil {
				return
			}
		}
		if err == io.EOF {
			fmt.Printf("flac streaming is over %d bytes", totalSent)
			time.Sleep(2 * time.Second)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte("complete"))
			if err != nil {
				return
			}
			break

		}

		totalSent += int64(n)

	}

}

func (c *Client) disconnect() {
	if c.isClosed {
		return
	}
	c.isClosed = true

	// Send close message
	err := c.conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
	if err != nil {
		return
	}

	err = c.conn.Close()
	if err != nil {
		return
	}
	close(c.write)

	log.Printf("Client %d disconnected\n", c.ID)
}
