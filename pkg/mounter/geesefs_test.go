package mounter

import (
	"slices"
	"testing"

	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

func TestValidateGeeseFSOptionsForms(t *testing.T) {
	want := []string{
		"--memory-limit", "1000",
		"--dir-mode", "0777",
		"--file-mode", "0666",
	}

	tests := []struct {
		name    string
		options []string
	}{
		{
			name: "separate values",
			options: []string{
				"--memory-limit", "1000",
				"--dir-mode", "0777",
				"--file-mode", "0666",
			},
		},
		{
			name: "equals values",
			options: []string{
				"--memory-limit=1000",
				"--dir-mode=0777",
				"--file-mode=0666",
			},
		},
		{
			name: "mixed forms",
			options: []string{
				"--memory-limit=1000",
				"--dir-mode", "0777",
				"--file-mode=0666",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateGeeseFSOptions(tt.options)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestValidateGeeseFSOptionsRejectsUnsafeFlags(t *testing.T) {
	flags := []struct {
		name  string
		value string
	}{
		{"iam", "auto"},
		{"iam-flavor", "gcp"},
		{"iam-url", "http://example.invalid"},
		{"iam-header", "test"},
		{"endpoint", "https://example.invalid"},
		{"pprof", "127.0.0.1:6060"},
		{"setuid", "0"},
		{"setgid", "0"},
		{"log-file", "/tmp/geesefs.log"},
		{"shared-config", "/tmp/config"},
		{"cache", "/tmp/cache"},
		{"no-systemd", "true"},
		{"unknown-option", "value"},
	}

	for _, flag := range flags {
		for _, prefix := range []string{"-", "--"} {
			name := prefix + flag.name
			forms := []struct {
				name string
				args []string
			}{
				{"bare", []string{name}},
				{"separate", []string{name, flag.value}},
				{"equals", []string{name + "=" + flag.value}},
			}

			for _, form := range forms {
				t.Run(name+"/"+form.name, func(t *testing.T) {
					// Precede the forbidden flag with a valid option.
					options := append(
						[]string{"--memory-limit", "1000"},
						form.args...,
					)

					got, err := ValidateGeeseFSOptions(options)
					if err == nil {
						t.Fatal("expected forbidden option to be rejected")
					}
					if got != nil {
						t.Fatalf("returned partial arguments on error: %q", got)
					}
				})
			}
		}
	}
}

func TestMountRejectsUnsafeOptions(t *testing.T) {
	m := &geesefsMounter{
		meta: &s3.FSMeta{
			BucketName:   "test-bucket",
			MountOptions: []string{"--setuid", "0"},
		},
	}

	err := m.Mount(t.TempDir(), "test-volume")
	if err == nil {
		t.Fatal("expected unsafe option to be rejected")
	}
	if got, want := err.Error(), `geesefs option "--setuid" is not allowed`; got != want {
		t.Fatalf("got error %q, want %q", got, want)
	}
}
