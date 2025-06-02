package influx

import (
	"context"
	"time"

	"micze.io/gama350/config"
	"micze.io/gama350/logger"
	"micze.io/gama350/model"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
)

type Writer interface {
	Write(data model.MeterData)
}

type Service struct {
	client   influxdb2.Client
	writeAPI influxdb2api.WriteAPIBlocking
}

func NewService() *Service {
	client := influxdb2.NewClient(config.Get("INFLUX_URL"), config.Get("INFLUX_TOKEN"))
	writeAPI := client.WriteAPIBlocking(config.Get("INFLUX_ORG"), config.Get("INFLUX_BUCKET"))
	logger.Logger.Println("Connected to InfluxDB")
	return &Service{
		client:   client,
		writeAPI: writeAPI,
	}
}

func (s *Service) Write(data model.MeterData) {
	var timestamp time.Time
	loc, _ := time.LoadLocation("Europe/Warsaw")
	logger.Logger.Printf("Device date time: %s", data.DeviceDateTime)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", data.DeviceDateTime, loc)
	if err != nil {
		logger.Logger.Printf("Error parsing device date time: %v", err)
		timestamp = time.Now()
	} else {
		timestamp = t.UTC()
	}

	point := influxdb2.NewPoint("energy_telemetry",
		map[string]string{
			"source": "meter",
			"media":  data.Media,
			"meter":  data.Meter,
			"name":   data.Name,
			"id":     data.ID,
			"device": data.Device,
		},
		map[string]interface{}{
			"current_power_consumption_kw":    data.CurrentPowerConsumptionKW,
			"current_power_production_kw":     data.CurrentPowerProductionKW,
			"total_energy_consumption_kwh":    data.TotalEnergyConsumptionKWH,
			"total_energy_consumption_t1_kwh": data.TotalEnergyConsumptionT1KWH,
			"total_energy_consumption_t2_kwh": data.TotalEnergyConsumptionT2KWH,
			"total_energy_consumption_t3_kwh": data.TotalEnergyConsumptionT3KWH,
			"total_energy_production_kwh":     data.TotalEnergyProductionKWH,
			"total_energy_production_t1_kwh":  data.TotalEnergyProductionT1KWH,
			"total_energy_production_t2_kwh":  data.TotalEnergyProductionT2KWH,
			"total_energy_production_t3_kwh":  data.TotalEnergyProductionT3KWH,
			"voltage_at_phase_1_v":            data.VoltagePhase1V,
			"voltage_at_phase_2_v":            data.VoltagePhase2V,
			"voltage_at_phase_3_v":            data.VoltagePhase3V,
			"rssi_dbm":                        data.RssiDBM,
		}, timestamp)

	if err := s.writeAPI.WritePoint(context.Background(), point); err != nil {
		logger.Logger.Printf("Influx write error: %v", err)
	} else {
		logger.Logger.Printf("Data written for meter %s", data.ID)
	}
}

func (s *Service) Client() influxdb2.Client {
	return s.client
}

func (s *Service) Close() {
	s.client.Close()
}
