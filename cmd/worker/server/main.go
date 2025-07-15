package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"net/http"
)

// 定义一个 upgrader，将 HTTP 请求升级为 WebSocket 请求
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// GoPBX Server
func main() {
	r := gin.Default()

	r.GET("/ws", handleConnection)
	port := ":8080"
	fmt.Printf("WebSocket server running at %s...\n", port)
	err := r.Run(port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

type Server struct {
	conn     *websocket.Conn
	peerConn *webrtc.PeerConnection
}

func handleConnection(c *gin.Context) {
	// 升级 HTTP 请求为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	connection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})

	defer connection.Close()

	//远程音视频轨道（track）被接收到时触发
	connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("Received track", track.ID(), track.StreamID())
	})

	sample, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypePCMU,
		Channels:  1,
		ClockRate: 8000,
	}, "audio", "pion")
	if err != nil {
		log.Println("Error creating track:", err)
		return
	}

	//向当前连接中添加一个媒体轨道（如音频或视频）
	connection.AddTrack(sample)

	offer, err := connection.CreateOffer(nil)
	if err != nil {
		log.Println("Error creating offer:", err)
	}
	err = connection.SetLocalDescription(offer)
	if err != nil {
		log.Println("Error setting local description:", err)
	}

	connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Println("ICE connection state changed:", connectionState)
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("Connection established")
		}
	})
	<-webrtc.GatheringCompletePromise(connection)
	// 循环接收消息并回复
	for {
		// 读取消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// 打印收到的消息
		fmt.Printf("Received message: %s\n", p)
		err = conn.WriteMessage(messageType, p)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}
