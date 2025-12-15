package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// MicrozooConfigProperties entspricht der Konfiguration aus der Java-Anwendung
type MicrozooConfigProperties struct {
	RequestDelay   time.Duration
	ResponseDelay  time.Duration
	UpstreamServices []string
	EntityCount    int
	PayloadSize    int
}

// BaseDto entspricht der Datenstruktur aus der Java-Anwendung
type BaseDto struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

var config MicrozooConfigProperties

func loadConfig() {
	viper.SetDefault("microzoo.requestDelay", "0ms")
	viper.SetDefault("microzoo.responseDelay", "0ms")
	viper.SetDefault("microzoo.entityCount", 1)
	viper.SetDefault("microzoo.payloadSize", 100)
	
	// Konfiguration aus Umgebungsvariablen laden
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MICROZOO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// RequestDelay
	reqDelayStr := viper.GetString("REQUESTDELAY")
	if reqDelayStr == "" {
		reqDelayStr = "0ms"
	}
	reqDelay, err := time.ParseDuration(reqDelayStr)
	if err != nil {
		log.Printf("WARN: Konnte RequestDelay nicht parsen: %v. Verwende 0ms.", err)
		reqDelay = 0
	}
	config.RequestDelay = reqDelay

	// ResponseDelay
	respDelayStr := viper.GetString("RESPONSEDELAY")
	if respDelayStr == "" {
		respDelayStr = "0ms"
	}
	respDelay, err := time.ParseDuration(respDelayStr)
	if err != nil {
		log.Printf("WARN: Konnte ResponseDelay nicht parsen: %v. Verwende 0ms.", err)
		respDelay = 0
	}
	config.ResponseDelay = respDelay

	// UpstreamServices
	upstreamStr := viper.GetString("UPSTREAMSERVICES")
	if upstreamStr != "" {
		config.UpstreamServices = strings.Split(upstreamStr, ",")
	} else {
		config.UpstreamServices = []string{}
	}

	// EntityCount
	entityCountStr := viper.GetString("ENTITYCOUNT")
	if entityCountStr != "" {
		config.EntityCount, err = strconv.Atoi(entityCountStr)
		if err != nil {
			log.Printf("WARN: Konnte EntityCount nicht parsen: %v. Verwende 1.", err)
			config.EntityCount = 1
		}
	} else {
		config.EntityCount = 1
	}

	// PayloadSize
	payloadSizeStr := viper.GetString("PAYLOADSIZE")
	if payloadSizeStr != "" {
		config.PayloadSize, err = strconv.Atoi(payloadSizeStr)
		if err != nil {
			log.Printf("WARN: Konnte PayloadSize nicht parsen: %v. Verwende 100.", err)
			config.PayloadSize = 100
		}
	} else {
		config.PayloadSize = 100
	}

	log.Printf("Konfiguration geladen: %+v", config)
}

func generateBaseDto(id int) BaseDto {
	payload := strings.Repeat("x", config.PayloadSize)
	return BaseDto{
		ID:      fmt.Sprintf("go-%d", id),
		Name:    fmt.Sprintf("Go Entity %d", id),
		Payload: payload,
	}
}

func getAll(c *gin.Context) {
	log.Println("Entered GET /api/base")
	time.Sleep(config.RequestDelay)

	// Simuliere die Logik aus BaseService.java
	// Da wir keine Datenbank haben, simulieren wir nur die "No-Database"-Logik und Upstream-Aufrufe

	// 1. Fall: Upstream-Services sind konfiguriert
	if len(config.UpstreamServices) > 0 {
		log.Println("Fetching entities from upstream services")
		var dtos []BaseDto
		
		// Hier müsste die Logik für FeignClients/HTTP-Aufrufe zu Upstream-Services implementiert werden.
		// Für diese Demonstration wird dies vereinfacht und nur die Struktur gezeigt.
		// In einer vollständigen Implementierung würde man hier HTTP-Clients verwenden.
		
		// Simuliere den Aufruf und die Aggregation
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Delegating call to %s/api/base", serviceURL)
			// Echter HTTP-Aufruf würde hier erfolgen
			// Für die Demo geben wir einfach ein Dummy-Ergebnis zurück
			dtos = append(dtos, BaseDto{
				ID: fmt.Sprintf("upstream-%s-1", serviceURL),
				Name: fmt.Sprintf("Upstream Entity from %s", serviceURL),
				Payload: strings.Repeat("y", config.PayloadSize),
			})
		}
		
		time.Sleep(config.ResponseDelay)
		log.Println("Exiting GET /api/base (Upstream)")
		c.JSON(http.StatusOK, dtos)
		return
	}

	// 2. Fall: Keine Datenbank, keine Upstream-Services (Generierung von Dummy-Daten)
	log.Println("Generating dummy entities")
	var dtos []BaseDto
	for i := 1; i <= config.EntityCount; i++ {
		dtos = append(dtos, generateBaseDto(i))
	}

	time.Sleep(config.ResponseDelay)
	log.Println("Exiting GET /api/base (Dummy)")
	c.JSON(http.StatusOK, dtos)
}

func create(c *gin.Context) {
	log.Println("Entered POST /api/base")
	time.Sleep(config.RequestDelay)

	var baseDto BaseDto
	if err := c.ShouldBindJSON(&baseDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simuliere die Logik aus BaseService.java
	// 1. Fall: Upstream-Services sind konfiguriert
	if len(config.UpstreamServices) > 0 {
		log.Printf("Posting dto with id %s to upstream services", baseDto.ID)
		
		// Hier müsste die Logik für FeignClients/HTTP-Aufrufe zu Upstream-Services implementiert werden.
		// Für diese Demonstration wird dies vereinfacht.
		
		// Simuliere den Aufruf und die Rückgabe
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Posting dto with id %s to service %s", baseDto.ID, serviceURL)
			// Echter HTTP-Aufruf würde hier erfolgen
		}
		
		time.Sleep(config.ResponseDelay)
		log.Println("Exiting POST /api/base (Upstream)")
		c.JSON(http.StatusCreated, baseDto)
		return
	}

	// 2. Fall: Keine Datenbank, keine Upstream-Services (einfache Rückgabe)
	time.Sleep(config.ResponseDelay)
	log.Println("Exiting POST /api/base (No-DB)")
	c.JSON(http.StatusCreated, baseDto)
}

func main() {
	loadConfig()

	// Gin im Release-Modus für weniger Log-Ausgabe
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health Check Endpunkt
	router.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// REST Endpunkte
	api := router.Group("/api/base")
	{
		api.GET("/", getAll)
		api.POST("/", create)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go Service gestartet auf Port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Konnte Server nicht starten: %v", err)
	}
}
