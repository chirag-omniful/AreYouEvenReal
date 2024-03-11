package main

import (
	_ "fmt"
	"log"
	"net"
	_ "os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Domain model for database
type Domain struct {
	gorm.Model
	Name        string
	HasMX       bool
	HasSPF      bool
	SPFRecord   string
	HasDMARC    bool
	DMARCRecord string
}

var db *gorm.DB

func main() {
	// Initialize the database
	initDB()

	// Setup Gin router
	router := gin.Default()

	// API endpoint for checking a domain
	router.POST("/checkDomain", func(c *gin.Context) {
		var domainRequest struct {
			Domain string `json:"domain" binding:"required"`
		}

		if err := c.ShouldBindJSON(&domainRequest); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		result := checkDomain(domainRequest.Domain)

		// Save the result to the database
		db.Create(&result)

		c.JSON(200, result)
	})

	// Run the server
	router.Run(":8080")
}

func initDB() {
	// Set up PostgreSQL connection details
	dsn := "host=localhost user=your_username password=your_password dbname=your_database port=5432 sslmode=disable"

	// Open a connection to PostgreSQL database
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// AutoMigrate the Domain model
	database.AutoMigrate(&Domain{})

	// Assign the database instance to the global variable
	db = database
}

func checkDomain(domain string) Domain {
	var result Domain

	var hasMX, hasSPF, hasDMARC bool
	var spfRecord, dmarcRecord string

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}

	if len(mxRecords) > 0 {
		hasMX = true
	}

	txtRecords, err := net.LookupTXT(domain)
	if err != nil {
		log.Printf("Error:%v\n", err)
	}

	for _, record := range txtRecords {
		if strings.HasPrefix(record, "v=spf1") {
			hasSPF = true
			spfRecord = record
			break
		}
	}

	dmarcRecords, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		log.Printf("ErrorL%v\n", err)
	}

	for _, record := range dmarcRecords {
		if strings.HasPrefix(record, "v=DMARC1") {
			hasDMARC = true
			dmarcRecord = record
			break
		}
	}

	result = Domain{
		Name:        domain,
		HasMX:       hasMX,
		HasSPF:      hasSPF,
		SPFRecord:   spfRecord,
		HasDMARC:    hasDMARC,
		DMARCRecord: dmarcRecord,
	}

	return result
}
