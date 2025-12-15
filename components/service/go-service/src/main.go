package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL-Treiber
	"github.com/spf13/viper"
)

// MicrozooConfigProperties entspricht der Konfiguration aus der Java-Anwendung
type MicrozooConfigProperties struct {
	RequestDelay     time.Duration
	ResponseDelay    time.Duration
	UpstreamServices []string
	EntityCount      int
	PayloadSize      int
	// Datenbank-Konfiguration
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
}

// BaseDto entspricht der Datenstruktur aus der Java-Anwendung
type BaseDto struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

var config MicrozooConfigProperties
var db *sql.DB

func loadConfig() {
	// ... (Unveränderte Konfigurationslogik) ...
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

	// Datenbank-Konfiguration
	config.DBHost = viper.GetString("DB_HOST")
	config.DBPort = viper.GetString("DB_PORT")
	config.DBName = viper.GetString("DB_NAME")
	config.DBUser = viper.GetString("DB_USER")
	config.DBPass = viper.GetString("DB_PASS")

	log.Printf("Konfiguration geladen: %+v", config)
}

func initDB() {
	if config.DBHost == "" {
		log.Println("INFO: Keine Datenbank-Konfiguration gefunden. Datenbank-Funktionalität deaktiviert.")
		return
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPass, config.DBName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("FEHLER: Konnte keine Verbindung zur Datenbank herstellen: %v", err)
	}

	// Testen der Verbindung
	err = db.Ping()
	if err != nil {
		log.Fatalf("FEHLER: Datenbank-Ping fehlgeschlagen: %v", err)
	}

	log.Println("INFO: Datenbankverbindung erfolgreich hergestellt.")

	// Tabelle erstellen, falls nicht vorhanden (analog zu Liquibase/JPA-Auto-Create)
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS base (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255),
		payload TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("FEHLER: Konnte Tabelle nicht erstellen: %v", err)
	}
	log.Println("INFO: Tabelle 'base' erstellt oder existiert bereits.")
}

func generateBaseDto(id int) BaseDto {
	payload := strings.Repeat("x", config.PayloadSize)
	return BaseDto{
		ID:      fmt.Sprintf("go-%d", id),
		Name:    fmt.Sprintf("Go Entity %d", id),
		Payload: payload,
	}
}

func isDBActive() bool {
	return db != nil
}

func getAllFromDB() ([]BaseDto, error) {
	log.Println("Fetching entities from database")
	rows, err := db.Query("SELECT id, name, payload FROM base")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dtos []BaseDto
	for rows.Next() {
		var dto BaseDto
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.Payload); err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func saveToDB(dto BaseDto) (BaseDto, error) {
	log.Printf("Saving entity with id %s in database", dto.ID)
	insertSQL := `INSERT INTO base (id, name, payload) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $2, payload = $3`
	_, err := db.Exec(insertSQL, dto.ID, dto.Name, dto.Payload)
	if err != nil {
		return BaseDto{}, err
	}
	return dto, nil
}

func getAll(c *gin.Context) {
	log.Println("Entered GET /api/base")
	time.Sleep(config.RequestDelay)

	var result []BaseDto
	var err error

	// 1. Fall: Datenbank ist aktiv
	if isDBActive() {
		result, err = getAllFromDB()
		if err != nil {
			log.Printf("FEHLER beim Abrufen aus DB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else if len(config.UpstreamServices) > 0 {
		// 2. Fall: Upstream-Services sind konfiguriert
		log.Println("Fetching entities from upstream services")
		var dtos []BaseDto
		
		// Simuliere den Aufruf und die Aggregation
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Delegating call to %s/api/base", serviceURL)
			// Echter HTTP-Aufruf würde hier erfolgen
			dtos = append(dtos, BaseDto{
				ID: fmt.Sprintf("upstream-%s-1", serviceURL),
				Name: fmt.Sprintf("Upstream Entity from %s", serviceURL),
				Payload: strings.Repeat("y", config.PayloadSize),
			})
		}
		result = dtos
	} else {
		// 3. Fall: Keine Datenbank, keine Upstream-Services (Generierung von Dummy-Daten)
		log.Println("Generating dummy entities")
		var dtos []BaseDto
		for i := 1; i <= config.EntityCount; i++ {
			dtos = append(dtos, generateBaseDto(i))
		}
		result = dtos
	}

	time.Sleep(config.ResponseDelay)
	log.Println("Exiting GET /api/base")
	c.JSON(http.StatusOK, result)
}

func create(c *gin.Context) {
	log.Println("Entered POST /api/base")
	time.Sleep(config.RequestDelay)

	var baseDto BaseDto
	if err := c.ShouldBindJSON(&baseDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result BaseDto
	var err error

	// 1. Fall: Datenbank ist aktiv
	if isDBActive() {
		result, err = saveToDB(baseDto)
		if err != nil {
			log.Printf("FEHLER beim Speichern in DB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else if len(config.UpstreamServices) > 0 {
		// 2. Fall: Upstream-Services sind konfiguriert
		log.Printf("Posting dto with id %s to upstream services", baseDto.ID)
		
		// Simuliere den Aufruf und die Rückgabe
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Posting dto with id %s to service %s", baseDto.ID, serviceURL)
			// Echter HTTP-Aufruf würde hier erfolgen
		}
		result = baseDto
	} else {
		// 3. Fall: Keine Datenbank, keine Upstream-Services (einfache Rückgabe)
		result = baseDto
	}

	time.Sleep(config.ResponseDelay)
	log.Println("Exiting POST /api/base")
	c.JSON(http.StatusCreated, result)
}

func main() {
	loadConfig()
	initDB() // Initialisiere die Datenbankverbindung

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
