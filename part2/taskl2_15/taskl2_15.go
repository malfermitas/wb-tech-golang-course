package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Command represents a single command with possible redirections
type Command struct {
	Args        []string
	RedirectIn  string
	RedirectOut string
	AppendOut   bool
}

// CommandGroup represents a group of commands with an operator
type CommandGroup struct {
	operator        string
	commands        []Command
	previousSuccess bool
}

func executeCommand(input []string) {
	if len(input) == 0 {
		return
	}

	// Expand environment variables
	expandedInput := expandEnvironmentVariables(input)

	// Parse command chaining
	subCommands := parseCommandChaining(expandedInput)

	previousSuccess := true // Track success across command groups

	for _, subCmd := range subCommands {
		subCmd.previousSuccess = previousSuccess
		success := executeSingleCommandGroup(subCmd)

		// Update previousSuccess for next group based on operator and result
		switch subCmd.operator {
		case "&&":
			previousSuccess = previousSuccess && success
		case "||":
			previousSuccess = previousSuccess || success
		default:
			previousSuccess = success
		}
	}
}

// expandEnvironmentVariables replaces $VAR with their values
func expandEnvironmentVariables(args []string) []string {
	var result []string
	for _, arg := range args {
		expanded := os.ExpandEnv(arg)
		result = append(result, expanded)
	}
	return result
}

// parseCommandChaining parses commands with && and || operators
func parseCommandChaining(input []string) []CommandGroup {
	var groups []CommandGroup

	currentCommands := make([]Command, 0)
	var currentOperator string

	i := 0
	for i < len(input) {
		token := input[i]

		if token == "&&" || token == "||" {
			if len(currentCommands) > 0 {
				groups = append(groups, CommandGroup{
					operator:        currentOperator,
					commands:        currentCommands,
					previousSuccess: true,
				})
				currentCommands = make([]Command, 0)
			}
			currentOperator = token
		} else {
			// Parse a complete command with possible redirections
			cmd, consumed := parseCommandWithRedirections(input[i:])
			currentCommands = append(currentCommands, cmd)
			i += consumed - 1 // -1 because we increment in the loop
		}
		i++
	}

	// Add the last group if not empty
	if len(currentCommands) > 0 {
		groups = append(groups, CommandGroup{
			operator:        currentOperator,
			commands:        currentCommands,
			previousSuccess: true,
		})
	}

	return groups
}

// parseCommandWithRedirections parses a single command with possible redirections
func parseCommandWithRedirections(tokens []string) (Command, int) {
	cmd := Command{Args: make([]string, 0)}
	consumed := 0

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		consumed = i + 1

		if token == "<" {
			// Input redirection
			if i+1 < len(tokens) {
				cmd.RedirectIn = tokens[i+1]
				consumed++
				i++ // Skip the next token
			}
		} else if token == ">" {
			// Output redirection (overwrite)
			if i+1 < len(tokens) {
				cmd.RedirectOut = tokens[i+1]
				cmd.AppendOut = false
				consumed++
				i++ // Skip the next token
			}
		} else if token == ">>" {
			// Output redirection (append)
			if i+1 < len(tokens) {
				cmd.RedirectOut = tokens[i+1]
				cmd.AppendOut = true
				consumed++
				i++ // Skip the next token
			}
		} else if token == "&&" || token == "||" {
			// Reached next operator, stop parsing this command
			consumed = i
			break
		} else {
			// Regular argument
			cmd.Args = append(cmd.Args, token)
		}
	}

	return cmd, consumed
}

// executeSingleCommandGroup executes a group of commands with operator logic
func executeSingleCommandGroup(group CommandGroup) bool {
	allSuccess := true

	for i, command := range group.commands {
		// Apply operator logic based on previous success state
		switch group.operator {
		case "&&":
			if i > 0 && !group.previousSuccess {
				return false // Don't execute if previous failed with &&
			}
		case "||":
			if i > 0 && group.previousSuccess {
				return true // Don't execute if previous succeeded with ||
			}
		}

		if err := executeBuiltIn(command); err != nil {
			if err.Error() != "unknown command" {
				fmt.Println(err)
				allSuccess = false
			} else {
				success := runExternalCommand(command)
				if !success {
					allSuccess = false
				}
			}
		}
	}

	return allSuccess
}

// runExternalCommand executes an external command with redirection support
func runExternalCommand(cmd Command) bool {
	if len(cmd.Args) == 0 {
		return false
	}

	execCmd := exec.Command(cmd.Args[0], cmd.Args[1:]...)

	// Handle input redirection
	if cmd.RedirectIn != "" {
		file, err := os.Open(cmd.RedirectIn)
		if err != nil {
			fmt.Printf("Error opening input file %s: %v\n", cmd.RedirectIn, err)
			return false
		}
		defer file.Close()
		execCmd.Stdin = file
	} else {
		execCmd.Stdin = os.Stdin
	}

	// Handle output redirection
	if cmd.RedirectOut != "" {
		var file *os.File
		var err error

		if cmd.AppendOut {
			file, err = os.OpenFile(cmd.RedirectOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			file, err = os.Create(cmd.RedirectOut)
		}

		if err != nil {
			fmt.Printf("Error opening output file %s: %v\n", cmd.RedirectOut, err)
			return false
		}
		defer file.Close()
		execCmd.Stdout = file
	} else {
		execCmd.Stdout = os.Stdout
	}

	// Always redirect stderr to the shell's stderr
	execCmd.Stderr = os.Stderr

	err := execCmd.Run()
	return err == nil
}

// executeBuiltIn executes built-in commands with redirection support
func executeBuiltIn(cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("empty command")
	}

	command := cmd.Args
	args := command[1:]

	switch command[0] {
	case "cd":
		if len(args) < 1 {
			return fmt.Errorf("expected path")
		}
		err := os.Chdir(args[0])
		if err != nil {
			return err
		}
	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		// Handle output redirection for pwd
		if cmd.RedirectOut != "" {
			var file *os.File
			var err error

			if cmd.AppendOut {
				file, err = os.OpenFile(cmd.RedirectOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			} else {
				file, err = os.Create(cmd.RedirectOut)
			}

			if err != nil {
				return fmt.Errorf("error opening output file %s: %v", cmd.RedirectOut, err)
			}
			defer file.Close()
			fmt.Fprintln(file, dir)
		} else {
			fmt.Println(dir)
		}
	case "echo":
		output := strings.Join(args, " ")

		// Handle output redirection for echo
		if cmd.RedirectOut != "" {
			var file *os.File
			var err error

			if cmd.AppendOut {
				file, err = os.OpenFile(cmd.RedirectOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			} else {
				file, err = os.Create(cmd.RedirectOut)
			}

			if err != nil {
				return fmt.Errorf("error opening output file %s: %v", cmd.RedirectOut, err)
			}
			defer file.Close()
			fmt.Fprintln(file, output)
		} else {
			fmt.Println(output)
		}
	case "kill":
		if len(args) < 1 {
			return fmt.Errorf("expected pid")
		}
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		// Use os.FindProcess and proc.Kill instead of syscall.Kill
		proc, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		err = proc.Kill()
		if err != nil {
			return err
		}
	case "ps":
		execCmd := exec.Command("ps", "aux")

		// Handle output redirection for ps
		if cmd.RedirectOut != "" {
			var file *os.File
			var err error

			if cmd.AppendOut {
				file, err = os.OpenFile(cmd.RedirectOut, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			} else {
				file, err = os.Create(cmd.RedirectOut)
			}

			if err != nil {
				return fmt.Errorf("error opening output file %s: %v", cmd.RedirectOut, err)
			}
			defer file.Close()
			execCmd.Stdout = file
		} else {
			execCmd.Stdout = os.Stdout
		}

		execCmd.Stderr = os.Stderr
		output, err := execCmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		} else {
			if cmd.RedirectOut != "" {
				// Output already written to file
			} else {
				fmt.Print(string(output))
			}
		}
	case "exit":
		os.Exit(0)
	default:
		return fmt.Errorf("unknown command: %s", command[0])
	}
	return nil
}

func setupSignalHandler() chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	return sig
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	sigChan := setupSignalHandler()

	// Handle signals in a separate goroutine
	go func() {
		for {
			<-sigChan
			fmt.Print("\nShell interrupted. Press Ctrl+D to exit.\n")
		}
	}()

	previousSuccess := true

	for {
		pwd, _ := os.Getwd()
		prompt := fmt.Sprintf("%s $ ", filepath.Base(pwd))
		fmt.Print(prompt)

		input, err := reader.ReadString('\n')
		if err != nil {
			// Handle EOF (Ctrl+D)
			if err.Error() == "EOF" {
				fmt.Println()
				break
			}
			fmt.Println("Error reading input:", err)
			continue
		}

		cmdLine := strings.TrimSpace(input)
		if len(cmdLine) == 0 {
			continue
		}

		// Split input into tokens, respecting quotes
		tokens := splitIntoTokens(cmdLine)

		// Parse and execute commands
		subCommands := parseCommandChaining(tokens)

		for _, subCmd := range subCommands {
			success := executeSingleCommandGroup(subCmd)
			// Update previousSuccess for next iteration
			switch subCmd.operator {
			case "&&":
				previousSuccess = previousSuccess && success
			case "||":
				previousSuccess = previousSuccess || success
			default:
				previousSuccess = success
			}
		}
	}
}

// splitIntoTokens splits a command line into tokens, respecting quotes
func splitIntoTokens(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range input {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '"', '\'':
			inQuotes = !inQuotes
			current.WriteRune(r) // Keep quotes in the token
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(r)
			} else {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			}
		case '&', '|', '<', '>':
			if inQuotes {
				current.WriteRune(r)
			} else {
				// Handle operators
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}

				// Check for >> or ||
				if len(tokens) > 0 {
					lastToken := tokens[len(tokens)-1]
					if (lastToken == ">" && r == '>') || (lastToken == "|" && r == '|') {
						tokens[len(tokens)-1] += string(r)
						continue
					}
				}
				tokens = append(tokens, string(r))
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
