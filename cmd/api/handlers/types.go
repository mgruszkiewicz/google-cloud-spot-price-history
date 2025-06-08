package handlers

type Machine struct {
	MachineType          string
	RegionName           string
	MinHourSpotPrice     float64
	MaxHourSpotPrice     float64
	HourSpotPrice        float64
	SpotHourPriceHistory []PriceData
}

type Params struct {
	MachineType string `query:"machine_type"`
	RegionName  string `query:"region_name"`
}

type MachineTypes struct {
	Data []string `json:"machine_types"`
}

type Region struct {
	Data []string `json:"regions"`
}

type Hello struct {
	TotalRecords int `json:"total_records"`
}

type PriceData struct {
	Price     float64 `json:"price"`
	Timestamp string  `json:"timestamp"`
}
