package constant

import "fmt"

const BaseUrl = "https://app.blockmesh.xyz/api"
const SpeedTestUrl = "https://api.speed.hetzner.de"

var LoginURL = fmt.Sprintf("%v/get_token", BaseUrl)
var UptimeURL = fmt.Sprintf("%v/report_uptime", BaseUrl)
var BandwidthURL = fmt.Sprintf("%v/submit_bandwidth", BaseUrl)
var TaskURL = fmt.Sprintf("%v/get_task", BaseUrl)
