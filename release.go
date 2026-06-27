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

	var tags []string

	for _, skin := range skins {
		if !SkinChanged(skin) {
			fmt.Printf("%-20s unchanged\n", skin)
			continue
		}

		tag, err := ReleaseSkin(skin)
		if err != nil {
			panic(err)
		}

		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		fmt.Println("No skin changes detected.")
		return
	}

	for _, tag := range tags {
		fmt.Printf("Pushing %s\n", tag)

		cmd := exec.Command("git", "push", "origin", tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			panic(err)
		}
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

func ReleaseSkin(skin string) (string, error) {
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

	// if err := exec.Command("git", "tag", tag).Run(); err != nil {
	// 	return "", err
	// }
	cmd := exec.Command("git", "tag", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return tag, nil
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

func SkinVersion(skin string) string {
	tag := latestTag(skin)

	if tag == "" {
		return "v0.0.0"
	}

	parts := strings.Split(tag, "/")
	return parts[len(parts)-1]
}

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
