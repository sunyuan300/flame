package v1

type _ResponseScrapeList struct {
	ResId   string   `json:"res_id"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

type _ResponseScrapeInfo struct {
	ResId   string      `json:"res_id"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type _ResponseRemoveScrape struct {
	ResId   string `json:"res_id"`
	Message string `json:"message"`
}

type _ResponseAddNodeScrape struct {
	ResId   string `json:"res_id"`
	Message string `json:"message"`
}

type _ResponseUpdateNodeScrape struct {
	ResId   string `json:"res_id"`
	Message string `json:"message"`
}

type _RequestUpdateNodeScrape struct {
	ScrapeInterval int `json:"scrape_interval"`
	ScrapeTimeout  int `json:"scrape_timeout"`
	// default /metrics
	MetricsPath string `json:"metrics_path"`
	// 它至少应该包含psa和exporter类型标注
	Labels map[string]string `json:"labels"`
}
