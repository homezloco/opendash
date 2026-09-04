package catalog

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/opendash-project/opendash/internal/models"
)

var envKeyPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Port reservation protects concurrent OpenDash installs from selecting the
// same transient localhost port. It does not eliminate the unavoidable race
// against other processes, but it removes the intra-process race.
var (
	portMu        sync.Mutex
	reservedPorts = make(map[int]struct{})
)

// ReserveHostPorts marks the chosen host ports as in-use by this process.
func ReserveHostPorts(ports map[int]int) {
	portMu.Lock()
	defer portMu.Unlock()
	for _, hp := range ports {
		reservedPorts[hp] = struct{}{}
	}
}

// ReleaseHostPorts releases a previous reservation.
func ReleaseHostPorts(ports map[int]int) {
	portMu.Lock()
	defer portMu.Unlock()
	for _, hp := range ports {
		delete(reservedPorts, hp)
	}
}

// isReserved reports whether a host port is already reserved by an in-flight
// OpenDash operation.
func isReserved(p int) bool {
	portMu.Lock()
	defer portMu.Unlock()
	_, ok := reservedPorts[p]
	return ok
}

func RenderComposeFiles(m *models.Manifest, plan *models.InstallPlan, hostPorts map[int]int, configValues map[string]string, installPath string) (composePath string, composeData, envData []byte, err error) {
	if err := os.MkdirAll(installPath, 0700); err != nil {
		return "", nil, nil, fmt.Errorf("create install directory: %w", err)
	}

	var envBuilder strings.Builder
	fmt.Fprintf(&envBuilder, "OPENDASH_APP_ID=%s\n", m.ID)
	fmt.Fprintf(&envBuilder, "OPENDASH_PROJECT_NAME=%s\n", plan.ProjectName)
	fmt.Fprintf(&envBuilder, "OPENDASH_INSTALL_PATH=%s\n", installPath)

	for key, value := range configValues {
		if !envKeyPattern.MatchString(key) {
			return "", nil, nil, fmt.Errorf("invalid env key %q", key)
		}
		value = strings.Join(strings.Split(value, "\n"), "\\n")
		fmt.Fprintf(&envBuilder, "%s=%s\n", key, value)
	}
	for _, sec := range m.Secrets {
		if !envKeyPattern.MatchString(sec.Name) {
			return "", nil, nil, fmt.Errorf("invalid secret env key %q", sec.Name)
		}
		v := configValues[sec.Name]
		v = strings.Join(strings.Split(v, "\n"), "\\n")
		fmt.Fprintf(&envBuilder, "%s=%s\n", sec.Name, v)
	}

	for _, vol := range plan.Volumes {
		key := volumeEnvKey(vol.Name)
		fmt.Fprintf(&envBuilder, "%s=%s\n", key, vol.HostPath)
	}
	for _, port := range plan.Ports {
		hp := hostPorts[port.ContainerPort]
		key := portEnvKey(port.ContainerPort)
		fmt.Fprintf(&envBuilder, "%s=%d\n", key, hp)
	}

	envData = []byte(envBuilder.String())

	if m.Compose.Inline != nil {
		composeData, err = renderInlineCompose(m, plan, hostPorts, configValues, installPath)
		if err != nil {
			return "", nil, nil, err
		}
		composePath = filepath.Join(installPath, "compose.json")
		return composePath, composeData, envData, nil
	}

	// Referenced file mode: copy file verbatim and rely on .env substitution.
	composePath = filepath.Join(installPath, filepath.Base(m.Compose.File))
	data, err := os.ReadFile(m.Compose.File)
	if err != nil {
		return "", nil, nil, fmt.Errorf("read referenced compose file: %w", err)
	}
	return composePath, data, envData, nil
}

func renderInlineCompose(m *models.Manifest, plan *models.InstallPlan, hostPorts map[int]int, configValues map[string]string, installPath string) ([]byte, error) {
	// Deep copy inline model to avoid mutating manifest.
	payload, err := json.Marshal(m.Compose.Inline)
	if err != nil {
		return nil, err
	}
	var inline models.ManifestInlineCompose
	if err := json.Unmarshal(payload, &inline); err != nil {
		return nil, err
	}

	// Determine target service for runtime bindings.
	target := m.Compose.MainService
	if target == "" {
		for name := range inline.Services {
			target = name
			break
		}
	}

	for name, svc := range inline.Services {
		if svc.ContainerName == "" {
			svc.ContainerName = projectNamePrefix + m.ID + "-" + name
		}
		if svc.Restart == "" {
			svc.Restart = "unless-stopped"
		}
		if svc.Environment == nil {
			svc.Environment = make(map[string]string)
		}
		for k, v := range configValues {
			svc.Environment[k] = v
		}
		for _, sec := range m.Secrets {
			if v, ok := configValues[sec.Name]; ok {
				svc.Environment[sec.Name] = v
			}
		}
		inline.Services[name] = svc
	}

	targetSvc, ok := inline.Services[target]
	if ok {
		// Assign generated port mappings for declared endpoints.
		portMap := make(map[int]string)
		for _, p := range targetSvc.Ports {
			hp, cp, _ := ParsePortMapping(p)
			portMap[cp] = p
			_ = hp
		}
		for _, ep := range plan.Ports {
			hp := hostPorts[ep.ContainerPort]
			portMap[ep.ContainerPort] = fmt.Sprintf("127.0.0.1:%d:%d", hp, ep.ContainerPort)
		}
		var ports []string
		for cp := range portMap {
			ports = append(ports, portMap[cp])
		}
		sort.Strings(ports) // deterministic
		targetSvc.Ports = ports

		// Bind storage volumes as named Docker volumes.
		for _, vol := range plan.Volumes {
			volumeName := vol.HostPath
			targetSvc.Volumes = replaceVolumeForMountPath(targetSvc.Volumes, vol.MountPath, fmt.Sprintf("%s:%s", volumeName, vol.MountPath))
		}
		inline.Services[target] = targetSvc
	}

	if inline.Volumes == nil {
		inline.Volumes = make(map[string]any)
	}
	for _, vol := range plan.Volumes {
		volumeName := vol.HostPath
		if _, ok := inline.Volumes[volumeName]; !ok {
			inline.Volumes[volumeName] = map[string]any{}
		}
	}

	output := map[string]any{
		"services": inline.Services,
	}
	if len(inline.Networks) > 0 {
		output["networks"] = inline.Networks
	}
	if len(inline.Volumes) > 0 {
		output["volumes"] = inline.Volumes
	}

	return json.MarshalIndent(output, "", "  ")
}

// replaceVolumeForMountPath removes any existing volume entry for mountPath
// (including short-form anonymous volumes) and appends the desired volume spec.
func replaceVolumeForMountPath(existing []string, mountPath, spec string) []string {
	var out []string
	for _, v := range existing {
		if containerPathFromVolumeSpec(v) != mountPath {
			out = append(out, v)
		}
	}
	return append(out, spec)
}

// containerPathFromVolumeSpec extracts the container-side path from a Compose
// volume string. It understands "name:/path", "host:/path", and "/path" forms.
func containerPathFromVolumeSpec(spec string) string {
	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 1:
		return parts[0]
	default:
		return parts[1]
	}
}

func volumeEnvKey(name string) string {
	return "OPENDASH_APP_VOLUME_" + strings.ToUpper(sanitizeEnvKey(name))
}

func portEnvKey(port int) string {
	return fmt.Sprintf("OPENDASH_APP_PORT_%d", port)
}

func sanitizeEnvKey(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func ValidateEnvKey(key string) error {
	if !envKeyPattern.MatchString(key) {
		return fmt.Errorf("invalid environment/secret key %q", key)
	}
	return nil
}

// HostPorts allocates a free localhost port for every endpoint that does not
// already have a fixed mapping in the inline compose model.
func HostPorts(m *models.Manifest, overrides map[int]int) (map[int]int, error) {
	ports := make(map[int]int)
	if overrides != nil {
		for cp, hp := range overrides {
			ports[cp] = hp
		}
	}
	if m.Compose.Inline != nil {
		for _, svc := range m.Compose.Inline.Services {
			for _, p := range svc.Ports {
				hp, cp, err := ParsePortMapping(p)
				if err != nil {
					return nil, err
				}
				if hp > 0 {
					ports[cp] = hp
				}
			}
		}
	}
	for _, ep := range m.Endpoints {
		if _, ok := ports[ep.Port]; ok {
			continue
		}
		hp, err := FindFreePort()
		if err != nil {
			return nil, err
		}
		ports[ep.Port] = hp
	}
	return ports, nil
}

// FindFreePort returns a currently unused localhost TCP port. Callers that
// intend to hand the port to Docker should reserve it with ReserveHostPorts
// to avoid intra-process races.
func FindFreePort() (int, error) {
	for attempts := 0; attempts < 20; attempts++ {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return 0, err
		}
		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			ln.Close()
			return 0, err
		}
		port, err := strconv.Atoi(portStr)
		ln.Close()
		if err != nil {
			return 0, err
		}
		if !isReserved(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("could not allocate a free localhost port")
}

func PortFromEnvName(key string) (int, bool) {
	if !strings.HasPrefix(key, "OPENDASH_APP_PORT_") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(key, "OPENDASH_APP_PORT_"))
	if err != nil {
		return 0, false
	}
	return n, true
}
