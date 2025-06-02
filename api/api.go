package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"micze.io/gama350/config"
	"micze.io/gama350/logger"
)

type APIServer struct {
	InfluxClient influxdb2.Client
	QueryAPI     api.QueryAPI
}

func NewAPIServer(client influxdb2.Client) *APIServer {
	return &APIServer{
		InfluxClient: client,
		QueryAPI:     client.QueryAPI(config.Get("INFLUX_ORG")),
	}
}

func (s *APIServer) Start() {
	r := gin.Default()
	r.GET("/readings/:id", s.getReadingsByID)
	r.GET("/readings/:id/last", s.getLastReadingByID)
	r.GET("/readings/:id/daily", s.getTodayEnergyByID)
	r.GET("/readings/:id/daily/:date", s.getEnergyByIDAndDate)
	go func() {
		if err := r.Run(":8080"); err != nil {
			logger.Logger.Fatalf("Failed to start API server: %v", err)
		}
	}()
}

func (s *APIServer) getReadingsByID(c *gin.Context) {
	id := c.Param("id")
	flux := "from(bucket: \"" + config.Get("INFLUX_BUCKET") + "\")" +
		" |> range(start: -30d)" +
		" |> filter(fn: (r) => r._measurement == \"energy_telemetry\" and r.id == \"" + id + "\")" +
		" |> pivot(rowKey:[\"_time\"], columnKey:[\"_field\"], valueColumn:\"_value\")"

	result, err := s.QueryAPI.Query(context.Background(), flux)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var readings []map[string]interface{}
	for result.Next() {
		readings = append(readings, result.Record().Values())
	}

	c.JSON(http.StatusOK, readings)
}

func (s *APIServer) getLastReadingByID(c *gin.Context) {
	id := c.Param("id")
	flux := "from(bucket: \"" + config.Get("INFLUX_BUCKET") + "\")" +
		" |> range(start: -30d)" +
		" |> filter(fn: (r) => r._measurement == \"energy_telemetry\" and r.id == \"" + id + "\")" +
		" |> sort(columns: [\"_time\"], desc: true)" +
		" |> limit(n: 1)" +
		" |> pivot(rowKey:[\"_time\"], columnKey:[\"_field\"], valueColumn:\"_value\")"

	result, err := s.QueryAPI.Query(context.Background(), flux)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Next() {
		c.JSON(http.StatusOK, result.Record().Values())
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "no reading found"})
}

// Returns daily production and consumption for current or specified date.
// Use GET /readings/{id}/daily for today's values in CET.
// Use GET /readings/{id}/daily/YYYY-MM-DD for a specific date.
func (s *APIServer) getTodayEnergyByID(c *gin.Context) {
	id := c.Param("id")
	today := time.Now().In(time.FixedZone("CET", 2*3600)).Format("2006-01-02")
	s.respondWithDailyEnergy(c, id, today)
}

func (s *APIServer) getEnergyByIDAndDate(c *gin.Context) {
	id := c.Param("id")
	date := c.Param("date")
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (expected YYYY-MM-DD)"})
		return
	}
	s.respondWithDailyEnergy(c, id, date)
}

func (s *APIServer) respondWithDailyEnergy(c *gin.Context, id string, date string) {
	loc := time.FixedZone("CET", 2*3600)
	start, _ := time.ParseInLocation("2006-01-02", date, loc)
	logger.Logger.Printf("start: %s", start)
	end := start.Add(24 * time.Hour)

	flux := fmt.Sprintf(`from(bucket: "%s")
        |> range(start: %s, stop: %s)
        |> filter(fn: (r) => r._measurement == "energy_telemetry" and r.id == "%s" and 
            (r._field == "total_energy_production_kwh" or r._field == "total_energy_consumption_kwh"))
        |> sort(columns: ["_time"])`,
		config.Get("INFLUX_BUCKET"),
		start.UTC().Format(time.RFC3339),
		end.UTC().Format(time.RFC3339),
		id,
	)

	result, err := s.QueryAPI.Query(context.Background(), flux)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type dataPoint struct {
		time  time.Time
		value float64
	}

	var firstProd, lastProd *dataPoint
	var firstCons, lastCons *dataPoint

	for result.Next() {
		record := result.Record()
		field := record.Field()
		value := record.Value().(float64)
		t := record.Time()

		switch field {
		case "total_energy_production_kwh":
			if firstProd == nil {
				firstProd = &dataPoint{t, value}
			}
			lastProd = &dataPoint{t, value}
		case "total_energy_consumption_kwh":
			if firstCons == nil {
				firstCons = &dataPoint{t, value}
			}
			lastCons = &dataPoint{t, value}
		}
	}

	if result.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Err().Error()})
		return
	}

	response := gin.H{
		"id":   id,
		"date": date,
	}

	if firstProd != nil && lastProd != nil {
		response["daily_production_kwh"] = lastProd.value - firstProd.value
	} else {
		response["daily_production_kwh"] = nil
	}

	if firstCons != nil && lastCons != nil {
		response["daily_consumption_kwh"] = lastCons.value - firstCons.value
	} else {
		response["daily_consumption_kwh"] = nil
	}

	c.JSON(http.StatusOK, response)
}
