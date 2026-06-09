package mqtt

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/entries"
	"gorm.io/gorm"
)

type EntryCreator interface {
	Create(entry *entries.Entry) error
}

type MQTTService struct {
	client   mqtt.Client
	db       *gorm.DB
	entrySvc EntryCreator
	cfg      *configs.MQTTConfig
	handlers map[string]mqtt.MessageHandler
}

func NewService(db *gorm.DB, entrySvc EntryCreator, cfg *configs.MQTTConfig) *MQTTService {
	return &MQTTService{
		db:       db,
		entrySvc: entrySvc,
		cfg:      cfg,
		handlers: make(map[string]mqtt.MessageHandler),
	}
}

func (s *MQTTService) Start() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.cfg.Broker)
	opts.SetClientID(s.cfg.ClientID + "-" + uuid.New().String())
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	if s.cfg.Username != "" {
		opts.SetUsername(s.cfg.Username)
		opts.SetPassword(s.cfg.Password)
	}

	opts.SetDefaultPublishHandler(s.defaultHandler)
	opts.SetConnectionLostHandler(s.connectionLostHandler)

	s.client = mqtt.NewClient(opts)
	token := s.client.Connect()
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	log.Printf("MQTT connected to %s as %s", s.cfg.Broker, s.cfg.ClientID)

	s.Subscribe("life/entry/create", s.handleEntryCreate)
	s.Subscribe("life/habit/complete", s.handleHabitComplete)
	s.Subscribe("life/goal/complete", s.handleGoalComplete)

	log.Println("MQTT subscribed to life/entry/create, life/habit/complete, life/goal/complete")

	return nil
}

func (s *MQTTService) Stop() error {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
		log.Println("MQTT disconnected")
	}
	return nil
}

func (s *MQTTService) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := s.client.Publish(topic, s.cfg.QoS, false, data)
	token.Wait()
	return token.Error()
}

func (s *MQTTService) Subscribe(topic string, handler mqtt.MessageHandler) error {
	s.handlers[topic] = handler
	token := s.client.Subscribe(topic, s.cfg.QoS, handler)
	token.Wait()
	return token.Error()
}

func (s *MQTTService) defaultHandler(client mqtt.Client, msg mqtt.Message) {
	s.handleMessage(msg.Topic(), msg.Payload())
}

func (s *MQTTService) connectionLostHandler(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
}

func (s *MQTTService) handleEntryCreate(client mqtt.Client, msg mqtt.Message) {
	s.handleMessage(msg.Topic(), msg.Payload())
}

func (s *MQTTService) handleHabitComplete(client mqtt.Client, msg mqtt.Message) {
	s.handleMessage(msg.Topic(), msg.Payload())
}

func (s *MQTTService) handleGoalComplete(client mqtt.Client, msg mqtt.Message) {
	s.handleMessage(msg.Topic(), msg.Payload())
}

func (s *MQTTService) handleMessage(topic string, payload []byte) {
	switch topic {
	case "life/entry/create":
		var req struct {
			Type       string      `json:"type"`
			Value      interface{} `json:"value"`
			Activities interface{} `json:"activities"`
			Note       string      `json:"note"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Printf("failed to parse life/entry/create payload: %v", err)
			return
		}
		if req.Type == "" {
			req.Type = "custom"
		}
		score := 0
		if v, ok := req.Value.(float64); ok {
			score = int(v)
		}
		var activitiesStr string
		if req.Activities != nil {
			switch v := req.Activities.(type) {
			case string:
				activitiesStr = v
			default:
				b, _ := json.Marshal(v)
				activitiesStr = string(b)
			}
		}
		now := time.Now().Truncate(24 * time.Hour)
		entry := &entries.Entry{
			EntryType:  req.Type,
			MoodScore:  &score,
			Activities: &activitiesStr,
			Source:     "mqtt",
			Visibility: entries.VisibilityPrivate,
			EntryDate:  now,
		}
		if req.Note != "" {
			entry.Description = &req.Note
		}
		if err := s.entrySvc.Create(entry); err != nil {
			log.Printf("failed to create entry from MQTT: %v", err)
		} else {
			log.Printf("entry created via MQTT: %s", entry.ID)
		}

	case "life/habit/complete":
		var req struct {
			HabitID string  `json:"habit_id"`
			Value   float64 `json:"value"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Printf("failed to parse life/habit/complete payload: %v", err)
			return
		}
		log.Printf("habit completed via MQTT: %s (value: %.2f)", req.HabitID, req.Value)

	case "life/goal/complete":
		var req struct {
			GoalID string  `json:"goal_id"`
			Value  float64 `json:"value"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			log.Printf("failed to parse life/goal/complete payload: %v", err)
			return
		}
		log.Printf("goal progress via MQTT: %s (value: %.2f)", req.GoalID, req.Value)

	default:
		log.Printf("unhandled MQTT topic: %s", topic)
	}
}

func (s *MQTTService) HandlePublish(c *gin.Context) {
	var req struct {
		Topic   string      `json:"topic" binding:"required"`
		Payload interface{} `json:"payload" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := s.Publish(req.Topic, req.Payload); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "published"})
}

func RegisterRoutes(rg *gin.RouterGroup, mqttSvc *MQTTService, mw gin.HandlerFunc) {
	mqtt := rg.Group("/mqtt").Use(mw)
	{
		mqtt.POST("/publish", mqttSvc.HandlePublish)
	}
}
