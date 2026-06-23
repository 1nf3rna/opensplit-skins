package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type SkinInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func main() {
	entries, err := os.ReadDir("skins")
	if err != nil {
		panic(err)
	}

	var updated bool

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skin := entry.Name()

		changed := SkinChanged(skin)

		if !changed {
			fmt.Printf("%-20s unchanged\n", skin)
			continue
		}

		if err := bumpSkin(skin); err != nil {
			panic(err)
		}

		updated = true
	}

	if !updated {
		fmt.Println("No skin changes detected.")
	}
}

func ReleaseSkin(skin string) error {
	current := SkinVersion(skin)

	next := bumpPatch(current)

	tag := fmt.Sprintf(
		"skin/%s/%s",
		skin,
		next,
	)

	cmd := exec.Command("git", "tag", tag)

	return cmd.Run()
}

func SkinVersion(skin string) string {
	tagPattern := fmt.Sprintf("skin/%s/*", skin)

	cmd := exec.Command(
		"git",
		"tag",
		"--list",
		tagPattern,
		"--sort=-version:refname",
	)

	out, err := cmd.Output()
	if err != nil {
		return "v0.0.0"
	}

	tags := strings.Fields(string(out))
	if len(tags) == 0 {
		return "v0.0.0"
	}

	parts := strings.Split(tags[0], "/")
	return parts[len(parts)-1]
}

func SkinChanged(skin string) bool {
	tag := latestTag(skin)

	if tag == "" {
		return true
	}

	cmd := exec.Command(
		"git",
		"diff",
		"--name-only",
		tag,
		"HEAD",
		"--",
		filepath.Join("skins", skin),
	)

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) != ""
}

func latestTag(skin string) string {
	pattern := fmt.Sprintf("skin/%s/*", skin)

	out, err := exec.Command(
		"git",
		"tag",
		"--list",
		pattern,
		"--sort=-version:refname",
	).Output()

	if err != nil {
		return ""
	}

	lines := strings.Fields(string(out))
	if len(lines) == 0 {
		return ""
	}

	return lines[0]
}

func bumpSkin(skin string) error {
	path := filepath.Join("skins", skin, "skin.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var info SkinInfo

	if err := json.Unmarshal(data, &info); err != nil {
		return err
	}

	oldVersion := info.Version
	newVersion := bumpPatch(oldVersion)

	info.Version = newVersion

	out, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return err
	}

	run("git", "add", path)

	msg := fmt.Sprintf(
		"Release skin %s v%s",
		skin,
		newVersion,
	)

	run("git", "commit", "-m", msg)

	tag := fmt.Sprintf(
		"skin/%s/v%s",
		skin,
		newVersion,
	)

	run("git", "tag", tag)

	fmt.Printf(
		"%-20s %s -> %s\n",
		skin,
		oldVersion,
		newVersion,
	)

	return nil
}

func bumpPatch(v string) string {
	parts := strings.Split(v, ".")

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	patch++

	return fmt.Sprintf(
		"%d.%d.%d",
		major,
		minor,
		patch,
	)
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
