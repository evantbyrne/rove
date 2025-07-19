package rove

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/stoewer/go-strcase"
)

type ServiceState struct {
	Command             []string
	Env                 []string
	Image               string
	Init                bool
	Mounts              []string
	Networks            []string
	Publish             []string
	Replicas            string
	Secrets             []string
	UpdateDelay         string
	UpdateFailureAction string
	UpdateOrder         string
	UpdateParallelism   string
	User                string
	WorkDir             string
}

func (new *ServiceState) Diff(old *ServiceState) (string, DiffStatus) {
	lines := make([]DiffLine, 0)
	maxLeft := 0
	res := make([]string, len(lines))
	status := DiffSame

	lines, status = diffSlices(lines, status, "command", old.Command, new.Command)
	lines, status = diffSlices(lines, status, "env", old.Env, new.Env)
	lines, status = diffString(lines, status, "image", old.Image, new.Image)
	lines, status = diffBool(lines, status, "init", old.Init, new.Init)
	lines, status = diffSlices(lines, status, "mounts", old.Mounts, new.Mounts)
	lines, status = diffSlices(lines, status, "network", old.Networks, new.Networks)
	lines, status = diffSlices(lines, status, "publish", old.Publish, new.Publish)
	lines, status = diffString(lines, status, "replicas", old.Replicas, new.Replicas)
	lines, status = diffSlices(lines, status, "secret", old.Secrets, new.Secrets)
	lines, status = diffString(lines, status, "update-delay", old.UpdateDelay, new.UpdateDelay)
	lines, status = diffString(lines, status, "update-failure-action", old.UpdateFailureAction, new.UpdateFailureAction)
	lines, status = diffString(lines, status, "update-order", old.UpdateOrder, new.UpdateOrder)
	lines, status = diffString(lines, status, "update-parallelism", old.UpdateParallelism, new.UpdateParallelism)
	lines, status = diffString(lines, status, "user", old.User, new.User)
	lines, status = diffString(lines, status, "workdir", old.WorkDir, new.WorkDir)

	for _, line := range lines {
		if len(line.Left) > maxLeft {
			maxLeft = len(line.Left)
		}
	}
	for _, line := range lines {
		res = append(res, line.StringPadded(maxLeft))
	}
	return strings.Join(res, "\n"), status
}

func formatStateMapKebab(state map[string]string) string {
	var out strings.Builder
	for i, key := range slices.Sorted(maps.Keys(state)) {
		if i != 0 {
			out.WriteString(",")
		}
		out.WriteString(fmt.Sprint(strcase.KebabCase(key), "=", state[key]))
	}
	return out.String()
}

func formatStateMount(mount DockerServiceMountJson) string {
	out := make([]string, 0)
	if mount.BindOptions.Propagation != "" {
		out = append(out, fmt.Sprint("bind-propagation=", mount.BindOptions.Propagation))
	}
	if mount.BindOptions.NonRecursive {
		out = append(out, "bind-nonrecursive=true")
	}
	if mount.Consistency != "" {
		out = append(out, fmt.Sprint("consistency=", mount.Consistency))
	}
	if mount.ReadOnly {
		out = append(out, "readonly=true")
	}
	if mount.Source != "" {
		out = append(out, fmt.Sprint("source=", mount.Source))
	}
	if mount.Target != "" {
		out = append(out, fmt.Sprint("target=", mount.Target))
	}
	if mount.TmpfsOptions.Mode != "" {
		out = append(out, fmt.Sprint("tmpfs-mode=", mount.TmpfsOptions.Mode))
	}
	if mount.TmpfsOptions.SizeBytes > 0 {
		out = append(out, fmt.Sprint("tmpfs-size=", strconv.FormatUint(mount.TmpfsOptions.SizeBytes, 10)))
	}
	if mount.Type != "" {
		out = append(out, fmt.Sprint("type=", mount.Type))
	}
	if mount.VolumeOptions.DriverConfig.Name != "" {
		out = append(out, fmt.Sprint("volume-driver=", mount.VolumeOptions.DriverConfig.Name))
	}
	if mount.VolumeOptions.NoCopy {
		out = append(out, "volume-nocopy=true")
	}
	if len(mount.VolumeOptions.Labels) > 0 {
		labels := make([]string, 0)
		for _, key := range slices.Sorted(maps.Keys(mount.VolumeOptions.Labels)) {
			labels = append(labels, fmt.Sprint(key, "=", mount.VolumeOptions.Labels[key]))
		}
		out = append(out, "volume-label="+strings.Join(labels, ","))
	}
	if len(mount.VolumeOptions.DriverConfig.Options) > 0 {
		labels := make([]string, 0)
		for _, key := range slices.Sorted(maps.Keys(mount.VolumeOptions.DriverConfig.Options)) {
			labels = append(labels, fmt.Sprint(key, "=", mount.VolumeOptions.DriverConfig.Options[key]))
		}
		out = append(out, "volume-opt="+strings.Join(labels, ","))
	}
	return strings.Join(out, ",")
}
