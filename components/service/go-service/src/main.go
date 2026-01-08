package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MicrozooConfigProperties mirrors the configuration from the Java application
type MicrozooConfigProperties struct {
	RequestDelay     time.Duration
	ResponseDelay    time.Duration
	UpstreamServices []string
	EntityCount      int
	PayloadSize      int
	// Database Configuration
	DBHost string // For PostgreSQL
	DBPort string // For PostgreSQL
	DBName string // For PostgreSQL
	DBUser string // For PostgreSQL
	DBPass string // For PostgreSQL
	// MongoDB Configuration
	MongoURI   string
	MongoDBName string
}

// BaseDto mirrors the data structure from the Java application
type BaseDto struct {
	ID      string `json:"id" bson:"_id"`
	Name    string `json:"name" bson:"name"`
	Payload string `json:"payload" bson:"payload"`
}

var config MicrozooConfigProperties
var sqlDB *sql.DB
var mongoClient *mongo.Client
var mongoCollection *mongo.Collection

func loadConfig() {
	// Load configuration from environment variables
	viper.SetDefault("microzoo.requestDelay", "0ms")
	viper.SetDefault("microzoo.responseDelay", "0ms")
	viper.SetDefault("microzoo.entityCount", 1)
	viper.SetDefault("microzoo.payloadSize", 100)
	
	// Load configuration from environment variables
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
		log.Printf("WARN: Could not parse RequestDelay: %v. Using 0ms.", err)
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
		log.Printf("WARN: Could not parse ResponseDelay: %v. Using 0ms.", err)
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
			log.Printf("WARN: Could not parse EntityCount: %v. Using 1.", err)
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
			log.Printf("WARN: Could not parse PayloadSize: %v. Using 100.", err)
			config.PayloadSize = 100
		}
	} else {
		config.PayloadSize = 100
	}

	// Database Configuration (PostgreSQL)
	config.DBHost = viper.GetString("DB_HOST")
	config.DBPort = viper.GetString("DB_PORT")
	config.DBName = viper.GetString("DB_NAME")
	config.DBUser = viper.GetString("DB_USER")
	config.DBPass = viper.GetString("DB_PASS")

	// MongoDB Configuration
	config.MongoURI = viper.GetString("MONGO_URI")
	config.MongoDBName = viper.GetString("MONGO_DBNAME")

	log.Printf("Configuration loaded: %+v", config)
}

func initDB() {
	// 1. MongoDB Initialization
	if config.MongoURI != "" && config.MongoDBName != "" {
		log.Println("INFO: MongoDB configuration found. Attempting connection...")
		clientOptions := options.Client().ApplyURI(config.MongoURI)
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		mongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("ERROR: Could not connect to MongoDB: %v", err)
		}

		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			log.Fatalf("ERROR: MongoDB ping failed: %v", err)
		}

		mongoCollection = mongoClient.Database(config.MongoDBName).Collection("base")
		log.Println("INFO: MongoDB connection established successfully.")
		return
	}

	// 2. PostgreSQL Initialization (Fallback)
	if config.DBHost != "" {
		log.Println("INFO: PostgreSQL configuration found. Attempting connection...")
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.DBHost, config.DBPort, config.DBUser, config.DBPass, config.DBName)

		var err error
		sqlDB, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("ERROR: Could not connect to database: %v", err)
		}

		// Test connection
		err = sqlDB.Ping()
		if err != nil {
			log.Fatalf("ERROR: Database ping failed: %v", err)
		}

		log.Println("INFO: PostgreSQL connection established successfully.")

		// Create table if not exists
		createTableSQL := `
		CREATE TABLE IF NOT EXISTS base (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255),
			payload TEXT
		);`
		_, err = sqlDB.Exec(createTableSQL)
		if err != nil {
			log.Fatalf("ERROR: Could not create table: %v", err)
		}
		log.Println("INFO: Table 'base' created or already exists.")
		return
	}

	log.Println("INFO: No database configuration found. Database functionality disabled.")
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
	return sqlDB != nil || mongoClient != nil
}

// --- SQL (PostgreSQL) Logic ---

func getAllFromSQL() ([]BaseDto, error) {
	log.Println("Fetching entities from PostgreSQL")
	rows, err := sqlDB.Query("SELECT id, name, payload FROM base")
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

func saveToSQL(dto BaseDto) (BaseDto, error) {
	log.Printf("Saving entity with id %s in PostgreSQL", dto.ID)
	insertSQL := `INSERT INTO base (id, name, payload) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $2, payload = $3`
	_, err := sqlDB.Exec(insertSQL, dto.ID, dto.Name, dto.Payload)
	if err != nil {
		return BaseDto{}, err
	}
	return dto, nil
}

// --- MongoDB Logic ---

func getAllFromMongo() ([]BaseDto, error) {
	log.Println("Fetching entities from MongoDB")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := mongoCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dtos []BaseDto
	if err = cursor.All(ctx, &dtos); err != nil {
		return nil, err
	}
	return dtos, nil
}

func saveToMongo(dto BaseDto) (BaseDto, error) {
	log.Printf("Saving entity with id %s in MongoDB", dto.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MongoDB uses upsert for saving/updating
	filter := bson.M{"_id": dto.ID}
	update := bson.M{"$set": dto}
	opts := options.Update().SetUpsert(true)

	_, err := mongoCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return BaseDto{}, err
	}
	return dto, nil
}

// --- Controller Logic ---

func getAll(c *gin.Context) {
	log.Println("Entered GET /api/base")
	time.Sleep(config.RequestDelay)

	var result []BaseDto
	var err error

	// Case 1: Database is active
	if isDBActive() {
		if mongoClient != nil {
			result, err = getAllFromMongo()
		} else if sqlDB != nil {
			result, err = getAllFromSQL()
		}

		if err != nil {
			log.Printf("ERROR fetching from DB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else if len(config.UpstreamServices) > 0 {
		// Case 2: Upstream services are configured
		log.Println("Fetching entities from upstream services")
		var dtos []BaseDto
		
		// Simulate call and aggregation
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Delegating call to %s/api/base", serviceURL)
			// Actual HTTP call would happen here
			dtos = append(dtos, BaseDto{
				ID: fmt.Sprintf("upstream-%s-1", serviceURL),
				Name: fmt.Sprintf("Upstream Entity from %s", serviceURL),
				Payload: strings.Repeat("y", config.PayloadSize),
			})
		}
		result = dtos
	} else {
		// Case 3: No database, no upstream services (generate dummy data)
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

	// Case 1: Database is active
	if isDBActive() {
		if mongoClient != nil {
			result, err = saveToMongo(baseDto)
		} else if sqlDB != nil {
			result, err = saveToSQL(baseDto)
		}

		if err != nil {
			log.Printf("ERROR saving to DB: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else if len(config.UpstreamServices) > 0 {
		// Case 2: Upstream services are configured
		log.Printf("Posting dto with id %s to upstream services", baseDto.ID)
		
		// Simulate call and return
		for _, serviceURL := range config.UpstreamServices {
			log.Printf("Posting dto with id %s to service %s", baseDto.ID, serviceURL)
			// Actual HTTP call would happen here
		}
		result = baseDto
	} else {
		// Case 3: No database, no upstream services (simple return)
		result = baseDto
	}

	time.Sleep(config.ResponseDelay)
	log.Println("Exiting POST /api/base")
	c.JSON(http.StatusCreated, result)
}

func main() {
	loadConfig()
	initDB() // Initialize database connection

	// Gin in release mode for less log output
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health Check Endpoint
	router.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// REST Endpoints
	api := router.Group("/api/base")
	{
		api.GET("/", getAll)
		api.POST("/", create)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go Service started on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
