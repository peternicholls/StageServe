// Package logs streams Apple Container logs for a project, with named diagnostics when a
// requested service is missing or unhealthy.
package logs

import (
	"context"
	"fmt"
	"io"

	coreruntime "github.com/peternicholls/stageserve/core/runtime"
)

// Streamer streams logs for one runtime service.
type Streamer struct {
	Runtime coreruntime.Manager
}

// Stream finds the container by service+project label and copies its logs to w.
func (s *Streamer) Stream(ctx context.Context, runtimeProject, service string, follow bool, w io.Writer) error {
	containers, err := s.Runtime.ListServices(ctx, runtimeProject)
	if err != nil {
		return err
	}
	var selected *coreruntime.Service
	for i := range containers {
		if containers[i].Service == service {
			selected = &containers[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("logs: no container found for service %q in project %q", service, runtimeProject)
	}
	stream, err := s.Runtime.ServiceLogs(ctx, selected.ID, follow)
	if err != nil {
		return err
	}
	defer stream.Close()
	_, err = io.Copy(w, stream)
	return err
}
