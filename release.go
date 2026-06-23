package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	skins, err := skinDirs()
	if err != nil {
		panic(err)
	}

	var updated bool

	for _, skin := range skins {
		if !SkinChanged(skin) {
			fmt.Printf("%-20s unchanged\n", skin)
			continue
		}

		if err := ReleaseSkin(skin); err != nil {
			panic(err)
		}

		updated = true
	}

	if !updated {
		fmt.Println("No skin changes detected.")
	}
}

func skinDirs() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var skins []string

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		skins = append(skins, name)
	}

	return skins, nil
}

func ReleaseSkin(skin string) error {
	current := SkinVersion(skin)

	var next string

	if current == "v0.0.0" {
		next = "1.0.0"
	} else {
		next = bumpPatch(current)
	}

	tag := fmt.Sprintf("%s/v%s", skin, next)

	fmt.Printf(
		"%-20s %s -> v%s\n",
		skin,
		current,
		next,
	)

	return exec.Command("git", "tag", tag).Run()
}

func SkinVersion(skin string) string {
	tag := latestTag(skin)

	if tag == "" {
		return "v0.0.0"
	}

	parts := strings.Split(tag, "/")
	return parts[len(parts)-1]
}

// func SkinVersion(skin string) string {
// 	tagPattern := fmt.Sprintf("skin/%s/*", skin)

// 	cmd := exec.Command(
// 		"git",
// 		"tag",
// 		"--list",
// 		tagPattern,
// 		"--sort=-version:refname",
// 	)

// 	out, err := cmd.Output()
// 	if err != nil {
// 		return "v0.0.0"
// 	}

// 	tags := strings.Fields(string(out))
// 	if len(tags) == 0 {
// 		return "v0.0.0"
// 	}

// 	parts := strings.Split(tags[0], "/")
// 	return parts[len(parts)-1]
// }

func SkinChanged(skin string) bool {
	tag := latestTag(skin)

	var cmd *exec.Cmd

	if tag == "" {
		cmd = exec.Command(
			"git",
			"log",
			"--oneline",
			"--",
			skin,
		)
	} else {
		cmd = exec.Command(
			"git",
			"log",
			"--oneline",
			fmt.Sprintf("%s..HEAD", tag),
			"--",
			skin,
		)
	}

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) != ""
}

func latestTag(skin string) string {
	pattern := fmt.Sprintf("%s/v*", skin)

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

	tags := strings.Fields(string(out))
	if len(tags) == 0 {
		return ""
	}

	return tags[0]
}

func bumpPatch(v string) string {
	v = strings.TrimPrefix(v, "v")

	parts := strings.Split(v, ".")

	if len(parts) != 3 {
		return "0.0.1"
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	patch++

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// func run(name string, args ...string) {
// 	cmd := exec.Command(name, args...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	if err := cmd.Run(); err != nil {
// 		panic(err)
// 	}
// }
