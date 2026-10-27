package container

// ContainerSpec describes a container to create.
type ContainerSpec struct {
	Name    string
	Image   string
	Env     map[string]string
	CapAdd  []string
	Devices []string // host paths, mapped to the same path in the container
	Ports   []PortBinding
	Mounts  []Mount
	Sysctls map[string]string
	Restart RestartPolicy
}

type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

// PortBinding publishes a container port on a host port.
type PortBinding struct {
	HostPort      uint16
	ContainerPort uint16
	Protocol      Protocol
}

// Mount binds a host directory to a path in the container.
type Mount struct {
	Source string
	Target string
}

type RestartPolicy string

const (
	RestartNo            RestartPolicy = "no"
	RestartUnlessStopped RestartPolicy = "unless-stopped"
)
