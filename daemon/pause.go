package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"fmt"
	"time"
	"github.com/docker/docker/container"
	"github.com/sirupsen/logrus"
)

// ContainerPause pauses a container
func (daemon *Daemon) ContainerPause(name string) error {
	logrus.Info("----------------------------containerPause in pause.go starts from ",int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond))
	ctr, err := daemon.GetContainer(name)
	if err != nil {
		return err
	}
	return daemon.containerPause(ctr)
}

// containerPause pauses the container execution without stopping the process.
// The execution can be resumed by calling containerUnpause.
func (daemon *Daemon) containerPause(container *container.Container) error {
	container.Lock()
	defer container.Unlock()

	// We cannot Pause the container which is not running
	if !container.Running {
		return errNotRunning(container.ID)
	}

	// We cannot Pause the container which is already paused
	if container.Paused {
		return errNotPaused(container.ID)
	}

	// We cannot Pause the container which is restarting
	if container.Restarting {
		return errContainerIsRestarting(container.ID)
	}

	if err := daemon.containerd.Pause(context.Background(), container.ID); err != nil {
		return fmt.Errorf("Cannot pause container %s: %s", container.ID, err)
	}

	container.Paused = true
	daemon.setStateCounter(container)
	daemon.updateHealthMonitor(container)
	daemon.LogContainerEvent(container, "pause")

	if err := container.CheckpointTo(daemon.containersReplica); err != nil {
		logrus.WithError(err).Warn("could not save container to disk")
	}
	logrus.Info("----------------------------containerPause in pause.go ends at ",int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond))
	return nil
}
