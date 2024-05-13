package flywatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/superfly/fly-go"
	bolt "go.etcd.io/bbolt"
)

const (
	updateInterval = 5 * time.Minute
)

type Controller struct {
	Db           *bolt.DB
	Logger       *slog.Logger
	Client       *http.Client
	APIToken     string
	Notify       <-chan struct{}
	ResyncPeriod time.Duration
}

type apiError struct {
	Error string `json:"error"`
}

func (c *Controller) Run() {
	c.Logger.Info("starting controller")

	resync := time.NewTimer(c.ResyncPeriod)
	defer resync.Stop()

	for {
		// Either we update every 5 minutes or when signaled from the API server
		select {
		case <-resync.C:
			c.Logger.Debug("resync period elapsed")
		case <-c.Notify:
			c.Logger.Debug("notified to resync")
		}

		c.sync()
	}
}

func (c *Controller) sync() {
	// TODO: Don't do all of this in a transaction?
	c.Db.View(func(tx *bolt.Tx) error {
		// Iterate over all buckets. The top le
		cursor := tx.Cursor()
		for bucket, _ := cursor.First(); bucket != nil; bucket, _ = cursor.Next() {
			appBucket := tx.Bucket(bucket)
			machinesBucket := appBucket.Bucket([]byte("machines"))

			cursor := appBucket.Cursor()
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				// Ignore buckets, they have nil values
				if v == nil {
					continue
				}

				c.Logger.Debug("handling deployment", "deployment", string(k))
				var md MachineDeployment
				err := json.Unmarshal(v, &md)
				if err != nil {
					c.Logger.Error("error unmarshaling machine deployment", "deployment", string(k), "error", err)
				}

				// TODO: Handle the case where we've updated our configuration, we need to update
				// all the machines

				// So, we have a set of replicas we want to create. We also know which machines we've already created
				// We go through each replica, compare to what we already know, then create or delete a machines if needed
				for _, replica := range md.Replicas {
					existingMachines := map[string]machine{}
					em := machinesBucket.Get([]byte(replica.Region))
					json.Unmarshal(em, &existingMachines)

					switch {
					// we don't know about this region, create all machines. no need to worry
					// about updates here
					case em == nil:
						c.Logger.Info(fmt.Sprintf("Creating %d machines in region %s", replica.Count, replica.Region))
						machines := map[string]machine{}
						for i := 0; i < replica.Count; i++ {
							// m, err := c.createMachine(md.App, createMachineRequest{
							// 	Region: replica.Region,
							// 	Config: &md.MachineConfig,
							// })
							// if err != nil {
							// 	c.Logger.Error("failed to create machine", "error", err)
							// }
							m := machine{}
							machines[m.ID] = m
						}
						// Store information in the bucket.
						// v, _ := json.Marshal(machines)
						// machinesBucket.Put([]byte(replica.Region), v)
					case len(existingMachines) == replica.Count:
						// check if they're all still in good shape. If we have a failed machine, then
						// we need to create a new one.
					case len(existingMachines) < replica.Count:
						// create machines
					case len(existingMachines) > replica.Count:
						// delete machines
					}
				}
			}
		}

		return nil
	})
}

func (c *Controller) createMachine(app string, request createMachineRequest) (*machine, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(request); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/apps/%s/machines", MachinesURL, app)

	req, _ := http.NewRequest(http.MethodPost, url, &b)
	req.Header.Set("authorization", c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK, resp.StatusCode == http.StatusCreated:
		var m machine
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			c.Logger.Error("failed to decode machine", "error", err)
			return nil, err
		}

		c.Logger.Info("successfully created machine", "machine_id", m.ID)
		return &m, nil
	case resp.StatusCode < http.StatusInternalServerError:
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("failed to decode API error (%d): %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to create machine (%d): %s", resp.StatusCode, apiErr.Error)
	default:
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
			c.Logger.Error("failed to create machine", "error", apiErr.Error, "status_code", resp.StatusCode)
		}
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}
}

type createMachineRequest struct {
	Region string             `json:"region"`
	Config *fly.MachineConfig `json:"config"`
}

type machine struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Region string `json:"region"`
	// InstanceID is unique for each version of the machine
	InstanceID string `json:"instance_id"`
	// PrivateIP is the internal 6PN address of the machine.
	PrivateIP string            `json:"private_ip"`
	Config    fly.MachineConfig `json:"config"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`

	Events []fly.MachineEvent `json:"events,omitempty"`
}
