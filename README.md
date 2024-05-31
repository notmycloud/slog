# slog
Golang SLog library for my packages


## JSONSchema
```go
package main

import "github.com/invopop/jsonschema"

func main() {
	print(generateSchema())
}

// User is used as a base to provide tests for comments.
type User struct {
	// Unique sequential identifier.
	ID int `json:"id" jsonschema:"required"`
	// Name of the user
	Name string `json:"name"`
}

func generateSchema() string {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/<ACCOUNT>/<REPO>", "./"); err != nil {
		// deal with error
	}
	return r.Reflect(&User{})
}
```