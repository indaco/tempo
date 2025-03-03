package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsEmbedded(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "File is embedded",
			path: "component/templ/component.templ.gotxt",
			want: true,
		},
		{
			name: "Directory is embedded",
			path: "component",
			want: true,
		},
		{
			name: "File is not embedded",
			path: "component/file.templ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmbedded(tt.path)
			if got != tt.want {
				t.Errorf("IsEmbedded(%q) = %v; want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestReadEmbeddedFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid embedded file",
			path:    "component/templ/component.templ.gotxt",
			wantErr: false,
		},
		{
			name:    "Non-existent file",
			path:    "nonexistent/file.templ",
			wantErr: true,
		},
		{
			name:    "Directory instead of file",
			path:    "component",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadEmbeddedFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadEmbeddedFile(%q) error = %v, wantErr = %v", tt.path, err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) == 0 {
				t.Errorf("ReadEmbeddedFile(%q) returned empty content; expected content", tt.path)
			}
		})
	}
}

func TestReadEmbeddedDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		want    []string // Expected file names
	}{
		{
			name:    "Valid embedded directory",
			path:    "component",
			wantErr: false,
			want:    []string{"assets", "templ"},
		},
		{
			name:    "Valid embedded directory",
			path:    "component-variant",
			wantErr: false,
			want:    []string{"assets", "name.templ.gotxt"},
		},
		{
			name:    "Non-existent directory",
			path:    "nonexistent",
			wantErr: true,
		},
		{
			name:    "File instead of directory",
			path:    "component/example.templ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadEmbeddedDir(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadEmbeddedDir(%q) error = %v, wantErr = %v", tt.path, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var gotFiles []string
				for _, entry := range got {
					gotFiles = append(gotFiles, entry.Name())
				}
				if !reflect.DeepEqual(gotFiles, tt.want) {
					t.Errorf("ReadEmbeddedDir(%q) = %v; want %v", tt.path, gotFiles, tt.want)
				}
			}
		})
	}
}

func TestCopyFileFromEmbed(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		destination string
		wantErr     bool
	}{
		{
			name:        "Copy valid embedded file",
			source:      "component/templ/component.templ.gotxt",
			destination: filepath.Join(os.TempDir(), "component.templ"),
			wantErr:     false,
		},
		{
			name:        "Copy non-existent file",
			source:      "nonexistent/file.templ",
			destination: filepath.Join(os.TempDir(), "invalid.templ"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CopyFileFromEmbed(tt.source, tt.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyFileFromEmbed(%q, %q) error = %v, wantErr = %v", tt.source, tt.destination, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(tt.destination); os.IsNotExist(err) {
					t.Errorf("Expected file %q to exist, but it does not", tt.destination)
				}

				// Cleanup
				os.Remove(tt.destination)
			}
		})
	}
}

func TestCopyDirFromEmbed(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		destination string
		wantErr     bool
	}{
		{
			name:        "Copy valid embedded directory",
			source:      "component/templ",
			destination: filepath.Join(os.TempDir(), "templ"),
			wantErr:     false,
		},
		{
			name:        "Copy non-existent directory",
			source:      "nonexistent/dir",
			destination: filepath.Join(os.TempDir(), "nonexistent"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CopyDirFromEmbed(tt.source, tt.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyDirFromEmbed(%q, %q) error = %v, wantErr = %v", tt.source, tt.destination, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify directory was created
				if _, err := os.Stat(tt.destination); os.IsNotExist(err) {
					t.Errorf("Expected directory %q to exist, but it does not", tt.destination)
				}

				// Cleanup
				os.RemoveAll(tt.destination)
			}
		})
	}
}
