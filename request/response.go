package request

type LoginResponse struct {
	APIToken string      `json:"api_token"`
	Message  interface{} `json:"message"`
}

type SpeedTestResult struct {
	DownloadSpeed float64 `json:"download_speed"`
	UploadSpeed   float64 `json:"upload_speed"`
	Latency       float64 `json:"latency"`
	City          string  `json:"city"`
	Country       string  `json:"country"`
	IP            string  `json:"ip"`
	ASN           string  `json:"asn"`
	Colo          string  `json:"colo"`
}

type GetIPResponse struct {
	IP string `json:"ip"`
}

type IpInformation struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}
