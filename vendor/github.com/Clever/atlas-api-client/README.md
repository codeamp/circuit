# atlas-api-client [![GoDoc](https://godoc.org/github.com/Clever/atlas-api-client/gen-go/client?status.png)](https://godoc.org/github.com/Clever/atlas-api-client/gen-go/client)

Go and JavaScript clients for MongoDB Atlas

Owned by eng-infra.

## Usage

``` go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Clever/atlas-api-client/digestauth"
	"github.com/Clever/atlas-api-client/gen-go/client"
)

func main() {
	atlasAPI := client.New("https://cloud.mongodb.com")
	digestT := digestauth.NewTransport(
		"usename",
		"apikey",
	)
	atlasAPI.SetTransport(&digestT)
  	clusters, err := atlasAPI.GetClusters(ctx, "groupID")
    // ...
}
```


- Run `make generate` to generate the code.
## Developing

- Update swagger.yml with your endpoints. See the [Swagger spec](http://swagger.io/specification/) for additional details on defining your swagger file.

- Run `make generate` to generate the code.
