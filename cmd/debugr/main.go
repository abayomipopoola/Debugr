package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/abayomipopoola/debugr/client"
	"github.com/abayomipopoola/debugr/utils"
)

var (
	debugMode   = flag.Bool("debug", false, "Enable debug mode")
	dryRun      = flag.Bool("dry-run", false, "Print suggested actions without executing")
	contextFile = flag.String("context", "", "Path to the file to use as context")
	contextDir  = flag.String("context-dir", "", "Path to the directory to use as context")
	apiKey      string
)

func init() {
	flag.Parse()
	apiKey = os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY is required")
	}
}

func main() {
	if len(flag.Args()) < 1 {
		log.Fatal("Usage: debugr [--context <file>] <prompt>")
		log.Fatal("Usage: debugr [--context-dir <directory>] <prompt>")
	}

	if *debugMode {
		utils.EnableDebug()
	}

	utils.Log("Initializing client with API key: %s", utils.MaskAPIKey(apiKey))

	c, err := client.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var fileContext *client.FileContext
	if *contextFile != "" {
		fileContext = loadSingleFileContext(*contextFile)
	} else if *contextDir != "" {
		fileContext = loadMultiFileContext(*contextDir)
	}

	prompt := strings.Join(flag.Args(), " ")
	utils.Log("Sending prompt: %s", prompt)

	actions, err := c.Prompt(prompt, fileContext)
	if err != nil {
		log.Fatalf("Failed to get actions: %v", err)
	}

	if len(actions) == 0 {
		log.Println("No actions suggested.")
		return
	}

	fmt.Println("Suggested actions:")
	for i, action := range actions {
		fmt.Printf("%d. ", i+1)
		switch action.Type {
		case "CREATE_FILE", "MODIFY_FILE":
			fmt.Printf("%s: %s\n", action.Type, action.Path)
			fmt.Printf("Content:-\n%s\n", action.Content)
		case "COMMAND":
			fmt.Printf("Execute:- %s\n", action.Content)
		case "EXPLANATION":
			fmt.Printf("Note:- %s\n", action.Content)
		}
	}

	if *dryRun {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for _, action := range actions {
		if action.Type == "EXPLANATION" {
			fmt.Println(action.Content)
			continue
		}

		fmt.Printf("Execute this action? (y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if response != "y" {
			continue
		}

		switch action.Type {
		case "CREATE_FILE", "MODIFY_FILE":
			err := utils.CreateOrModifyFile(action.Path, action.Content)
			if err != nil {
				log.Printf("Failed to %s: %v", strings.ToLower(action.Type), err)
			} else {
				fmt.Printf("File %s: %s\n", strings.ToLower(strings.TrimSuffix(action.Type, "_FILE")), action.Path)
			}
		case "COMMAND":
			cmd := exec.Command("sh", "-c", action.Content)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Printf("Failed to execute command: %v\nCommand was: %s", err, action.Content)
			}
		}
	}
}

func loadSingleFileContext(filePath string) *client.FileContext {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read context file: %v", err)
	}
	return &client.FileContext{
		Files: []client.File{
			{
				Path:     filePath,
				Content:  string(content),
				Language: utils.GetFileLanguage(filePath),
			},
		},
	}
}

func loadMultiFileContext(dirPath string) *client.FileContext {
	var files []client.File
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files = append(files, client.File{
				Path:     path,
				Content:  string(content),
				Language: utils.GetFileLanguage(path),
			})
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to read context directory: %v", err)
	}
	return &client.FileContext{Files: files}
}
