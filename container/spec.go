package container

// ContainerSpec describes a container to create.
type ContainerSpec struct {
	Name    string
	Image   string
	Labels  map[string]string
	Env     map[string]string
	CapAdd  []string
	Devices []string // host paths, mapped to the same path in the container
	Ports   []PortBinding
	Mounts  []Mount
	Sysctls map[string]string
	Restart RestartPolicy
}

// Container describes a container that exists on the host.
type Container struct {
	ID     string
	Name   string
	Labels map[string]string
	Env    map[string]string
}

type NetProto string

const (
	TCP NetProto = "tcp"
	UDP NetProto = "udp"
)

// PortBinding publishes a container port on a host port.
type PortBinding struct {
	HostPort      uint16
	ContainerPort uint16
	Protocol      NetProto
}

type Mount struct {
	HostPath      string
	ContainerPath string
}

type RestartPolicy string

const (
	RestartNo            RestartPolicy = "no"
	RestartUnlessStopped RestartPolicy = "unless-stopped"
)
