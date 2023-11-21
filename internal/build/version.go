package build

import "time"

var BuildTime = time.Now().UTC().Format(time.RFC3339)
var Version = "unknown"
var GitCommit = "unknown"
