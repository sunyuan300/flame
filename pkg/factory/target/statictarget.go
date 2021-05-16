package target

type StaticTarget struct {
	//Target address:port
	Targets []string `json:"targets"`
}
