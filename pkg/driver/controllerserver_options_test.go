package driver

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateVolumeRejectsUnsafeGeeseFSOptions(t *testing.T) {
	for _, mounterType := range []string{"geesefs", "", "unknown"} {
		for _, options := range []string{
			"--setuid 0",
			"--setuid=0",
			"-setuid=0",
			`"--setuid" "0"`,
		} {
			t.Run(mounterType+"/"+options, func(t *testing.T) {
				cs := &controllerServer{}
				req := &csi.CreateVolumeRequest{
					Name: "test-volume",
					Parameters: map[string]string{
						"mounter": mounterType,
						"options": options,
					},
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessType: &csi.VolumeCapability_Mount{
								Mount: &csi.VolumeCapability_MountVolume{},
							},
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
							},
						},
					},
					// No secrets: validation must precede S3 client creation.
				}

				resp, err := cs.CreateVolume(context.Background(), req)
				if status.Code(err) != codes.InvalidArgument {
					t.Fatalf("expected InvalidArgument, got %v", err)
				}
				if got := status.Convert(err).Message(); got !=
					`geesefs option "--setuid" is not allowed` &&
					got != `geesefs option "-setuid" is not allowed` {
					t.Fatalf("unexpected error message: %q", got)
				}
				if resp != nil {
					t.Fatalf("expected no response, got %v", resp)
				}
			})
		}
	}
}
