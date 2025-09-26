package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
)

// Constants
const (
	AppName    = "arni"
	BinDir     = "./bin"
	PkgDir     = "./cmd/main.go"
	ExtLDFlags = "-framework Foundation -framework Metal -framework MetalKit -lggml-metal -lggml-blas"
)

// Patterns
const (
	WhisperBuildDirPatter     = "%s/build_go"
	WhisperIncludePathPattern = "%s/include:%s/ggml/include"
	WhisperLibraryPathPattern = "%s/src:%s/ggml/src:%s/ggml/src/ggml-blas:%s/ggml/src/ggml-metal"
)

var (
	WhisperRoot            string
	WhisperBuildDir        string
	WhisperIncludePath     string
	WhisperLibraryPath     string
	GGMLMetalPathResources string
)

var Commands map[string]func([]string) error

func main() {

	if os.Args[1] == "whisper" {
		fmt.Println("==> Configure & build whisper.cpp (static lib)")
		WhisperRoot := os.Args[2]
		if isExist, err := isFolderExists(WhisperRoot); err != nil {
			log.Fatal(err)
		} else if !isExist {
			log.Fatalf("Folder %s does not exist", WhisperRoot)
		}

		WhisperBuildDir := fmt.Sprintf(WhisperBuildDirPatter, WhisperRoot)

		whisperConfigCmd := exec.Command(
			"cmake", "-S",
			WhisperRoot, "-B",
			WhisperBuildDir,
			"-DCMAKE_BUILD_TYPE=Release",
			"-DBUILD_SHARED_LIBS=OFF")
		whisperBuildCmd := exec.Command("cmake", "--build", WhisperBuildDir, "--target", "whisper")

		whisperConfigCmd.Stdout, whisperConfigCmd.Stderr = os.Stdout, os.Stderr
		whisperBuildCmd.Stdout, whisperBuildCmd.Stderr = os.Stdout, os.Stderr

		err := whisperConfigCmd.Run()
		if err != nil {
			log.Fatalf("Whisper configure cmd err : %s", err)
		}

		err = whisperBuildCmd.Run()
		if err != nil {
			log.Fatalf("Whisper build cmd err : %s", err)
		}
		return
	}

	if os.Args[1] == "build" {
		WhisperRoot := os.Args[2]
		if isExist, err := isFolderExists(WhisperRoot); err != nil {
			log.Fatal(err)
		} else if !isExist {
			log.Fatalf("Folder %s does not exist", WhisperRoot)
		}

		WhisperBuildDir := fmt.Sprintf(WhisperBuildDirPatter, WhisperRoot)
		WhisperIncludePath := fmt.Sprintf(WhisperIncludePathPattern, WhisperRoot)
		WhisperLibraryPath := fmt.Sprintf(WhisperLibraryPathPattern, WhisperBuildDir, WhisperBuildDir, WhisperBuildDir, WhisperBuildDir)
		GGMLMetalPathResources := WhisperRoot

		goBuildCmd := exec.Command("go",
			"build",
			"-ldflags",
			fmt.Sprintf("-extldflags '%s'", ExtLDFlags),
			"-o",
			AppName,
			PkgDir)

		goBuildCmd.Env = append(os.Environ(),
			fmt.Sprintf("C_INCLUDE_PATH=%s", WhisperIncludePath),
			fmt.Sprintf("LIBRARY_PATH=%s", WhisperLibraryPath),
			fmt.Sprintf("GGML_METAL_PATH_RESOURCES=%s", GGMLMetalPathResources))

		goBuildCmd.Stdout, goBuildCmd.Stderr = os.Stdout, os.Stderr
		fmt.Println(goBuildCmd.String())
		err := goBuildCmd.Run()
		if err != nil {
			log.Fatalf("Go build err : %s", err)
		}
	}
}

func isFolderExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
