package vein

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Vein defines an injection point for fraglet code
type Vein struct {
	Name       string   `yaml:"name"`                 // Vein name (e.g., "python", "c")
	Container  string   `yaml:"container"`            // Container image (required)
	Extensions []string `yaml:"extensions,omitempty"` // File extensions that map to this vein (e.g., [".py"])
}

// VeinRegistry manages available veins
type VeinRegistry struct {
	veins map[string]*Vein
}

// NewVeinRegistry creates an empty registry
func NewVeinRegistry() *VeinRegistry {
	return &VeinRegistry{
		veins: make(map[string]*Vein),
	}
}

// Add adds a vein to the registry
// Returns an error if a vein with the same name already exists (prevents clobbering)
func (r *VeinRegistry) Add(vein *Vein) error {
	if vein.Name == "" {
		return fmt.Errorf("vein name is required")
	}
	if vein.Container == "" {
		return fmt.Errorf("vein container is required")
	}
	if _, exists := r.veins[vein.Name]; exists {
		return fmt.Errorf("duplicate vein name: %s", vein.Name)
	}
	r.veins[vein.Name] = vein
	return nil
}

// Get retrieves a vein by name
func (r *VeinRegistry) Get(name string) (*Vein, bool) {
	vein, ok := r.veins[name]
	return vein, ok
}

// ContainerImage returns the container image for this vein, resolving the tag
// via FRAGLET_VEINS_FORCE_TAG or FRAGLET_VEIN_TAG_DISCOVERY_ORDER.
func (v *Vein) ContainerImage() string {
	return ResolveImageTag(v.Container)
}

var imageExistsCache sync.Map

func replaceTag(container, tag string) string {
	lastColon := strings.LastIndex(container, ":")
	if lastColon == -1 {
		return container + ":" + tag
	}
	return container[:lastColon] + ":" + tag
}

func imageExistsLocally(image string) bool {
	if cached, ok := imageExistsCache.Load(image); ok {
		return cached.(bool)
	}
	cmd := exec.Command("docker", "image", "inspect", "--format", ".", image)
	err := cmd.Run()
	exists := err == nil
	imageExistsCache.Store(image, exists)
	return exists
}

// ResolveImageTag picks the right tag for a container image.
//
// Priority:
//  1. FRAGLET_VEINS_FORCE_TAG — hard override, replaces tag unconditionally
//  2. FRAGLET_VEIN_TAG_DISCOVERY_ORDER — comma-separated tags to try in order;
//     uses the first tag whose image exists locally, falls back to the last tag
//  3. The container string as-is from veins.yml
func ResolveImageTag(container string) string {
	if tag := os.Getenv("FRAGLET_VEINS_FORCE_TAG"); tag != "" {
		return replaceTag(container, tag)
	}

	if order := os.Getenv("FRAGLET_VEIN_TAG_DISCOVERY_ORDER"); order != "" {
		tags := strings.Split(order, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			candidate := replaceTag(container, tag)
			if imageExistsLocally(candidate) {
				return candidate
			}
		}
		if last := strings.TrimSpace(tags[len(tags)-1]); last != "" {
			return replaceTag(container, last)
		}
	}

	return container
}

// ApplyForceTag is kept for backward compatibility; delegates to ResolveImageTag.
func ApplyForceTag(container string) string {
	return ResolveImageTag(container)
}

// List returns all vein names
func (r *VeinRegistry) List() []string {
	names := make([]string, 0, len(r.veins))
	for name := range r.veins {
		names = append(names, name)
	}
	return names
}

// ResolveImageDigest returns the image reference with an immutable digest when available
// (e.g. ghcr.io/foo/python:3.12 -> ghcr.io/foo/python@sha256:...). Used for persisted
// fraglet artifacts so re-execution uses the same image. If the image cannot be resolved
// to a digest (e.g. local-only image), returns the original image reference unchanged.
func ResolveImageDigest(ctx context.Context, image string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{index .RepoDigests 0}}", image)
	out, err := cmd.Output()
	if err != nil {
		return image, nil // return as-is so save still works
	}
	digestRef := strings.TrimSpace(string(out))
	if digestRef == "" || digestRef == "<no value>" {
		return image, nil
	}
	return digestRef, nil
}
