package cmd

import (
	"fmt"
	"time"

	docker "github.com/ddn0/go-dockerclient"
	"github.com/ddn0/peanut/image"
	"github.com/ddn0/peanut/logwriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup unused docker resources",
	RunE:  runCleanup,
}

func runCleanup(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	timeout := time.After(viper.GetDuration("timeout"))
	done := make(chan error)
	go func() {
		defer close(done)
		lw := logwriter.NewColorWriter("")
		defer lw.Flush()
		opt := cleanupOpt{
			RemoveRunning: viper.GetBool("remove-running"),
		}
		res, err := cleanup(opt)
		if err != nil {
			done <- err
		}
		for _, v := range res.Volumes {
			fmt.Println("Removed volume ", v.Name)
		}
		for _, c := range res.Containers {
			fmt.Println("Removed container ", c.ID)
		}
		for _, i := range res.Images {
			fmt.Println("Removed image ", i.ID)
		}
	}()

	select {
	case <-timeout:
		return nil
	case err := <-done:
		return err
	}
}

type cleanupOpt struct {
	RemoveRunning bool
}

type cleanupResults struct {
	Volumes    []docker.Volume
	Images     []docker.APIImages
	Containers []*docker.Container
}

func cleanup(opt cleanupOpt) (*cleanupResults, error) {
	var err error

	cs, e := removeContainers(removeContainersOpt{
		RemoveRunning: opt.RemoveRunning,
	})
	if e != nil {
		err = e
	}

	vols, e := removeDanglingVolumes()
	if e != nil {
		err = e
	}

	var images []docker.APIImages

	is1, e := removeDanglingImages()
	if e != nil {
		err = e
	}
	images = append(images, is1...)

	return &cleanupResults{
		Volumes:    vols,
		Images:     images,
		Containers: cs,
	}, err
}

// Remove dangling volumes. These are volumes not referenced by any running
// containers. Return volumes removed (even if there was an error).
//
// Warning: this will remove any unreferenced data containers.
func removeDanglingVolumes() ([]docker.Volume, error) {
	var removed []docker.Volume

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return removed, err
	}

	vols, err := client.ListVolumes(docker.ListVolumesOptions{
		Filters: map[string][]string{
			"dangling": []string{"true"},
		},
	})
	if err != nil {
		return removed, err
	}
	for _, vol := range vols {
		if err := client.RemoveVolume(vol.Name); err != nil {
			return removed, err
		}
		removed = append(removed, vol)
	}
	return removed, nil
}

func removeImage(filters map[string][]string, keepTags []string) ([]docker.APIImages, error) {
	keep := make(map[string]bool)
	for _, t := range keepTags {
		keep[t] = true
	}
	toKeep := func(img docker.APIImages) bool {
		for _, t := range img.RepoTags {
			_, _, tag := image.ParseImageName(t)
			if keep[tag] {
				return true
			}
		}
		return false
	}

	var removed []docker.APIImages

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return removed, err
	}

	seen := make(map[string]bool)
	for {
		imgs, err := client.ListImages(docker.ListImagesOptions{Filters: filters})
		if err != nil {
			return removed, err
		}

		any := false
		for _, img := range imgs {
			if seen[img.ID] {
				continue
			}
			if toKeep(img) {
				seen[img.ID] = true
				continue
			}

			if err := client.RemoveImageExtended(img.ID, docker.RemoveImageOptions{Force: true}); err != nil {
				return removed, err
			}
			removed = append(removed, img)
			seen[img.ID] = true
			any = true
		}

		if !any {
			return removed, nil
		}
	}
}

// Remove dangling images. These are images that have been superseded by new
// versions. Return images removed.
func removeDanglingImages() ([]docker.APIImages, error) {
	return removeImage(map[string][]string{
		"dangling": []string{"true"},
	}, nil)
}

type removeContainersOpt struct {
	RemoveRunning bool // Remove running containers too
}

// Remove containers. Return containers removed.
func removeContainers(opt removeContainersOpt) ([]*docker.Container, error) {
	var removed []*docker.Container

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return removed, err
	}

	cs, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return removed, err
	}

	for _, c := range cs {
		if c, e := client.InspectContainer(c.ID); e != nil {
			err = e
			continue
		} else if !opt.RemoveRunning && (c.State.Running || c.State.Paused) {
			continue
		} else if e := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true}); e != nil {
			err = e
		} else {
			removed = append(removed, c)
		}
	}

	return removed, err
}
func init() {
	c := cleanupCmd
	flags := c.Flags()

	RootCmd.AddCommand(c)
	flags.Bool("remove-running", false, "Try to remove running containers too")
	flags.Duration("timeout", 30*time.Second, "Duration to wait until declaring failure")
}
