package model

type MeterData struct {
	Media                       string  `json:"media"`
	Meter                       string  `json:"meter"`
	Name                        string  `json:"name"`
	ID                          string  `json:"id"`
	CurrentPowerConsumptionKW   float64 `json:"current_power_consumption_kw"`
	CurrentPowerProductionKW    float64 `json:"current_power_production_kw"`
	TotalEnergyConsumptionKWH   float64 `json:"total_energy_consumption_kwh"`
	TotalEnergyConsumptionT1KWH float64 `json:"total_energy_consumption_tariff_1_kwh"`
	TotalEnergyConsumptionT2KWH float64 `json:"total_energy_consumption_tariff_2_kwh"`
	TotalEnergyConsumptionT3KWH float64 `json:"total_energy_consumption_tariff_3_kwh"`
	TotalEnergyProductionKWH    float64 `json:"total_energy_production_kwh"`
	TotalEnergyProductionT1KWH  float64 `json:"total_energy_production_tariff_1_kwh"`
	TotalEnergyProductionT2KWH  float64 `json:"total_energy_production_tariff_2_kwh"`
	TotalEnergyProductionT3KWH  float64 `json:"total_energy_production_tariff_3_kwh"`
	VoltagePhase1V              float64 `json:"voltage_at_phase_1_v"`
	VoltagePhase2V              float64 `json:"voltage_at_phase_2_v"`
	VoltagePhase3V              float64 `json:"voltage_at_phase_3_v"`
	DeviceDateTime              string  `json:"device_date_time"`
	Device                      string  `json:"device"`
	RssiDBM                     int     `json:"rssi_dbm"`
}
