package flywatch

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/superfly/fly-go"
	bolt "go.etcd.io/bbolt"
)

const (
	MachinesURL               = "https://api.machines.dev"
	defaultDeploymentStrategy = "rolling"
)

type Flywatch struct {
	Logger *slog.Logger
	Db     *bolt.DB
}

func (f *Flywatch) Run() {
	// Setup API server
	mux := http.NewServeMux()
	mux.HandleFunc("/schedule", f.schedule())
	f.Logger.Info("starting Flywatch API server")
	err := http.ListenAndServe(":8080", mux)
	f.Logger.Error("server error", "error", err)
}

func (f *Flywatch) schedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Ensure it's a post method

		var config Config
		_, err := toml.NewDecoder(r.Body).Decode(&config)
		if err != nil {
			f.Logger.With("error", err).Error("failed to decode configuration")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f.Logger.Info("Received new deployment", "app", config.App, "name", config.Name)

		// We have a valid configuration, we'll store in the DB in the way that makes
		// sense for the machine controller
		err = f.Db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(config.App))
			if err != nil {
				return err
			}

			// Create a machine configuration
			mc := fly.MachineConfig{
				Env:      map[string]string{},
				Metadata: map[string]string{},
			}

			// Image
			mc.Image = config.Build.Image

			// Guest
			for _, vm := range config.Compute {
				guest := fly.MachineGuest{}
				if vm.CPUKind != "" {
					guest.CPUKind = vm.CPUKind
				}
				if vm.MemoryMB > 0 {
					guest.MemoryMB = vm.MemoryMB
				}
				if vm.CPUs > 0 {
					guest.CPUs = vm.CPUs
				}

				mc.Guest = &guest
			}

			// Services
			for _, service := range config.Services {
				svc := fly.MachineService{
					Protocol:           service.Protocol,
					InternalPort:       service.InternalPort,
					Autostop:           service.AutoStopMachines,
					Autostart:          service.AutoStartMachines,
					MinMachinesRunning: service.MinMachinesRunning,
					Ports:              service.Ports,
					Concurrency:        service.Concurrency,
				}

				mc.Services = append(mc.Services, svc)
			}

			// Mounts
			// TODO: We actually need to query for valid volumes here to find
			// the right mount
			for _, mount := range config.Mounts {
				mnt := fly.MachineMount{
					Encrypted: true,
					Path:      mount.Destination,
					Name:      mount.Source,
				}

				if mount.InitialSize > 0 {
					mnt.SizeGb = mount.InitialSize
				}

				mc.Mounts = append(mc.Mounts, mnt)
			}

			// Restart
			restart := fly.MachineRestart{
				Policy:     fly.MachineRestartPolicyOnFailure,
				MaxRetries: 3,
			}
			mc.Restart = &restart

			md := MachineDeployment{
				App:                config.App,
				MachineConfig:      mc,
				Replicas:           config.Replicas,
				DeploymentStrategy: defaultDeploymentStrategy,
			}

			if config.Deploy != nil {
				if s := config.Deploy.Strategy; s != "" {
					md.DeploymentStrategy = s
				}
			}

			value, err := json.Marshal(md)
			if err != nil {
				return err
			}
			b.Put([]byte(config.Name), value)

			return nil
		})

		if err != nil {
			f.Logger.Error("error storing deployment", "error", err)
		}
	}
}

type Config struct {
	Name     string    `toml:"name"`
	App      string    `toml:"app"`
	Build    Build     `toml:"build"`
	Deploy   *Deploy   `toml:"deploy"`
	Compute  []Compute `toml:"vm"`
	Services []Service `toml:"services"`
	Mounts   []Mount   `toml:"mounts"`
	Replicas []Replica `toml:"replicas"`
}

type Service struct {
	Protocol           string                         `json:"protocol,omitempty" toml:"protocol"`
	InternalPort       int                            `json:"internal_port,omitempty" toml:"internal_port"`
	AutoStopMachines   *bool                          `json:"auto_stop_machines,omitempty" toml:"auto_stop_machines"`
	AutoStartMachines  *bool                          `json:"auto_start_machines,omitempty" toml:"auto_start_machines"`
	MinMachinesRunning *int                           `json:"min_machines_running,omitempty" toml:"min_machines_running,omitempty"`
	Ports              []fly.MachinePort              `json:"ports,omitempty" toml:"ports"`
	Concurrency        *fly.MachineServiceConcurrency `json:"concurrency,omitempty" toml:"concurrency"`
}

type Mount struct {
	Source      string `toml:"source,omitempty" json:"source,omitempty"`
	Destination string `toml:"destination,omitempty" json:"destination,omitempty"`
	InitialSize int    `toml:"initial_size,omitempty" json:"initial_size,omitempty"`
}

type Build struct {
	Image string `toml:"image,omitempty" json:"image,omitempty"`
}

type Deploy struct {
	Strategy string `toml:"strategy,omitempty"`
}

type Compute struct {
	*fly.MachineGuest `toml:",inline" json:",inline"`
}

type Replica struct {
	Region string `toml:"region"`
	Count  int    `toml:"count"`
}

type MachineDeployment struct {
	App                string            `json:"string"`
	MachineConfig      fly.MachineConfig `json:"machine_config"`
	Replicas           []Replica         `json:"replicas"`
	DeploymentStrategy string            `json:"deployment_strategy"`
}
