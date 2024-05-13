package flywatch

// import (
// 	"context"
// 	"log/slog"

// 	"github.com/samber/lo"
// 	"github.com/superfly/fly-go"
// 	"github.com/superfly/fly-go/flaps"
// )

// const url = "https://api.machines.dev"

// type Flywatch struct {
// 	/// The name of the Fly.io application
// 	App string
// }

// type Machines struct {
// 	MachineConfig      fly.MachineConfig `json:"machine_config"`
// 	Replicas           map[string]int    `json:"replicas"`
// 	DeploymentStrategy string            `json:"deployment_strategy"`
// }

// type MachineController struct {
// 	queue <-chan Machines

// 	client flaps.Client
// 	app    string
// 	apiUrl string
// 	logger *slog.Logger
// }

// type Volumes struct {
// 	Volume   fly.CreateVolumeRequest `json:"volume_config"`
// 	Replicas map[string]int          `json:"replicas"`
// }

// type VolumeController struct {
// 	queue <-chan Volumes

// 	client flaps.Client
// 	app    string
// 	apiUrl string
// 	logger *slog.Logger
// }

// func (mc *MachineController) run(ctx context.Context) {
// 	for item := range mc.queue {
// 		existing, err := mc.client.GetMachines(ctx)
// 		if err != nil {
// 		}

// 		byRegion := lo.GroupBy(existing, func(item fly.Volume) string {
// 			return item.Region
// 		})

// 		for region, replicas := range item.Replicas {
// 			// Figure out if we need to create or delete volumes
// 			existing := byRegion[region]
// 			num := replicas - len(existing)
// 			switch {
// 			case num == 0:
// 				// do nothing
// 			case num > 0:
// 				// create machines
// 			case num < 0:
// 				// delete machines
// 			}
// 		}
// 	}
// }

// func (vc *VolumeController) run(ctx context.Context) {
// 	for item := range vc.queue {
// 		existing, err := vc.client.GetVolumes(ctx)
// 		if err != nil {
// 		}

// 		byRegion := lo.GroupBy(existing, func(item fly.Volume) string {
// 			return item.Region
// 		})

// 		for region, replicas := range item.Replicas {
// 			// Figure out if we need to create or delete volumes
// 			existing := byRegion[region]
// 			num := replicas - len(existing)
// 			switch {
// 			case num == 0:
// 				// do nothing
// 			case num > 0:
// 				// create volumes
// 			case num < 0:
// 				// delete volumes
// 			}
// 		}
// 	}
// }
