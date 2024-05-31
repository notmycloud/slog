package nmcslog

import (
	"bytes"
	"log/slog"
	"reflect"
	"testing"

	"github.com/invopop/jsonschema"
)

// func TestDebugLogger(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want *slog.Logger
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := DebugLogger(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DebugLogger() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestFileOutput_GetPath(t *testing.T) {
	type fields struct {
		OutputBase  OutputBase
		Path        string
		Rotate      Rotate
		loggingFile string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fo := &FileOutput{
				OutputBase:  tt.fields.OutputBase,
				Path:        tt.fields.Path,
				Rotate:      tt.fields.Rotate,
				loggingFile: tt.fields.loggingFile,
			}
			if got := fo.GetPath(); got != tt.want {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileOutput_JSONSchemaExtend(t *testing.T) {
	type fields struct {
		OutputBase  OutputBase
		Path        string
		Rotate      Rotate
		loggingFile string
	}
	type args struct {
		schema *jsonschema.Schema
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fo := &FileOutput{
				OutputBase:  tt.fields.OutputBase,
				Path:        tt.fields.Path,
				Rotate:      tt.fields.Rotate,
				loggingFile: tt.fields.loggingFile,
			}
			fo.JSONSchemaExtend(tt.args.schema)
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name string
		want *slog.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Logger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Logger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputFormat_FromString(t *testing.T) {
	type args struct {
		format string
	}
	tests := []struct {
		name    string
		of      OutputFormat
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.of.FromString(tt.args.format); (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutputFormat_Handler(t *testing.T) {
	type args struct {
		opts *slog.HandlerOptions
	}
	tests := []struct {
		name  string
		of    OutputFormat
		args  args
		wantW string
		want  slog.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got := tt.of.Handler(w, tt.args.opts)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Handler() gotW = %v, want %v", gotW, tt.wantW)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputFormat_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		of      OutputFormat
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.of.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutputFormat_UnmarshalTOML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		of      OutputFormat
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.of.UnmarshalTOML(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTOML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutputFormat_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		of      OutputFormat
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.of.UnmarshalYAML(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDebugLogger(t *testing.T) {
	type args struct {
		l *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *slog.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetDebugLogger(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetDebugLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDefaultLogger(t *testing.T) {
	type args struct {
		l *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *slog.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetDefaultLogger(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetDefaultLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}
