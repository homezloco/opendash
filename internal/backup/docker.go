package backup

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// DockerArchiver runs a fixed, non-shell command. Manifest values are validated
// as Docker volume names by Plan; no manifest-provided executable is accepted.
type DockerArchiver struct{ Image string }

func (d DockerArchiver) ArchiveVolume(ctx context.Context, volume, destination string) error {
	image := d.Image
	if image == "" {
		image = "alpine:3.20"
	}
	dir, name := filepath.Dir(destination), filepath.Base(destination)
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--network", "none", "--read-only", "-v", volume+":/source:ro", "-v", dir+":/backup", image, "tar", "-cf", "/backup/"+name, "-C", "/source", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker volume archiver: %w: %s", err, string(output))
	}
	return nil
}
