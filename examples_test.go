package nmcslog_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/invopop/jsonschema"
	nmcslog "github.com/notmycloud/slog"
	"golang.org/x/mod/module"
)

// Example function names
// Godoc uses a naming convention to associate an example function with a package-level identifier.
//
// func ExampleFoo()     // documents the Foo function or type
// func ExampleBar_Qux() // documents the Qux method of type Bar
// func Example()        // documents the package as a whole
// Following this convention, godoc displays the ExampleString example alongside the documentation for the String function.
//
// Multiple examples can be provided for a given identifier by using a suffix beginning with an underscore followed by a lowercase letter. Each of these examples documents the String function:
//
// func ExampleString()
// func ExampleString_second()
// func ExampleString_third()

func Example_generateSchema() {
	r := new(jsonschema.Reflector)
	print("Adding Go Struct Comments.\n")
	bInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Could not retrieve Build Info!")
	}
	// bInfo.Path should report the Go Module name such as github.com/ACCOUNT/REPOSITORY
	if err := module.CheckPath(bInfo.Path); err != nil {
		print("Invalid Module Path\n")
		print(bInfo.Path)
		print("\n")
		panic(err)
	}
	print("Getting comments from: " + bInfo.Path)
	if err := r.AddGoComments(bInfo.Path, "./"); err != nil {
		panic(err)
	}

	print("Only marking required as requested in jsonschema tags.\n")
	// Only tags explicitly marked as required instead of any that don't have `json:,omitempty`.
	r.RequiredFromJSONSchemaTags = true

	print("Analyzing the struct tree...\n")
	schema := r.Reflect(&nmcslog.Config{})

	print("Formatting the JSON Schema...\n")
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}

	print("Opening the Schema file.\n")
	f, _ := os.OpenFile("./config.schema.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("Closing Schema File", err)
			panic(err)
		}
	}()

	print("Writing the Schema file.\n")
	if _, err := f.WriteString(string(data)); err != nil {
		panic(err)
	}

	print("Schema Generated Successfully!\n")
	// NOTE: print() does not count as output!
	// NOTE: The // Output: comment must be "alone"

	// Output:
}
