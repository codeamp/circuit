package constants

import "os"

var StagingEnvironment = os.Getenv("STAGING_ENVIRONMENT")
var ProductionEnvironment = os.Getenv("PRODUCTION_ENVIRONMENT")
