package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestExpandEnvironmentVariables(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No variables",
			input:    []string{"echo", "hello"},
			expected: []string{"echo", "hello"},
		},
		{
			name:     "Single variable",
			input:    []string{"echo", "$TEST_VAR"},
			expected: []string{"echo", "test_value"},
		},
		{
			name:     "Multiple variables",
			input:    []string{"echo", "$TEST_VAR", "and", "$TEST_VAR"},
			expected: []string{"echo", "test_value", "and", "test_value"},
		},
		{
			name:     "Undefined variable",
			input:    []string{"echo", "$UNDEFINED_VAR"},
			expected: []string{"echo", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvironmentVariables(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("At index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestSplitIntoTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple command",
			input:    "ls -l",
			expected: []string{"ls", "-l"},
		},
		{
			name:     "Command with quotes",
			input:    `echo "hello world"`,
			expected: []string{`echo`, `"hello world"`},
		},
		{
			name:     "Command with single quotes",
			input:    `echo 'hello world'`,
			expected: []string{`echo`, `'hello world'`},
		},
		{
			name:     "Command with redirection",
			input:    "echo hello > output.txt",
			expected: []string{"echo", "hello", ">", "output.txt"},
		},
		{
			name:     "Command with append redirection",
			input:    "echo hello >> output.txt",
			expected: []string{"echo", "hello", ">>", "output.txt"},
		},
		{
			name:     "Command with input redirection",
			input:    "cat < input.txt",
			expected: []string{"cat", "<", "input.txt"},
		},
		{
			name:     "Command with escaped characters",
			input:    `echo hello\ world`,
			expected: []string{`echo`, `hello world`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitIntoTokens(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("At index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestParseCommandChaining(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []CommandGroup
	}{
		{
			name:  "Single command",
			input: []string{"ls", "-l"},
			expected: []CommandGroup{
				{
					operator: "",
					commands: []Command{
						{Args: []string{"ls", "-l"}},
					},
					previousSuccess: true,
				},
			},
		},
		{
			name:  "Two commands with &&",
			input: []string{"ls", "&&", "echo", "done"},
			expected: []CommandGroup{
				{
					operator: "",
					commands: []Command{
						{Args: []string{"ls"}},
					},
					previousSuccess: true,
				},
				{
					operator: "&&",
					commands: []Command{
						{Args: []string{"echo", "done"}},
					},
					previousSuccess: true,
				},
			},
		},
		{
			name:  "Two commands with ||",
			input: []string{"ls", "||", "echo", "failed"},
			expected: []CommandGroup{
				{
					operator: "",
					commands: []Command{
						{Args: []string{"ls"}},
					},
					previousSuccess: true,
				},
				{
					operator: "||",
					commands: []Command{
						{Args: []string{"echo", "failed"}},
					},
					previousSuccess: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommandChaining(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d groups, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %+v", tt.expected)
				t.Errorf("Got: %+v", result)
				return
			}

			for i, group := range result {
				expectedGroup := tt.expected[i]
				if group.operator != expectedGroup.operator {
					t.Errorf("Group %d: expected operator %q, got %q", i, expectedGroup.operator, group.operator)
				}

				if len(group.commands) != len(expectedGroup.commands) {
					t.Errorf("Group %d: expected %d commands, got %d", i, len(expectedGroup.commands), len(group.commands))
					continue
				}

				for j, cmd := range group.commands {
					expectedCmd := expectedGroup.commands[j]
					if len(cmd.Args) != len(expectedCmd.Args) {
						t.Errorf("Command %d,%d: expected %d args, got %d", i, j, len(expectedCmd.Args), len(cmd.Args))
						continue
					}

					for k, arg := range cmd.Args {
						if arg != expectedCmd.Args[k] {
							t.Errorf("Command %d,%d arg %d: expected %q, got %q", i, j, k, expectedCmd.Args[k], arg)
						}
					}

					// Проверяем поля перенаправления
					if cmd.RedirectIn != expectedCmd.RedirectIn {
						t.Errorf("Command %d,%d RedirectIn: expected %q, got %q", i, j, expectedCmd.RedirectIn, cmd.RedirectIn)
					}
					if cmd.RedirectOut != expectedCmd.RedirectOut {
						t.Errorf("Command %d,%d RedirectOut: expected %q, got %q", i, j, expectedCmd.RedirectOut, cmd.RedirectOut)
					}
					if cmd.AppendOut != expectedCmd.AppendOut {
						t.Errorf("Command %d,%d AppendOut: expected %t, got %t", i, j, expectedCmd.AppendOut, cmd.AppendOut)
					}
				}
			}
		})
	}
}

func TestParseCommandWithRedirections(t *testing.T) {
	tests := []struct {
		name             string
		input            []string
		expectedCmd      Command
		expectedConsumed int
	}{
		{
			name:  "Simple command",
			input: []string{"ls", "-l"},
			expectedCmd: Command{
				Args: []string{"ls", "-l"},
			},
			expectedConsumed: 2,
		},
		{
			name:  "Command with output redirection",
			input: []string{"echo", "hello", ">", "output.txt"},
			expectedCmd: Command{
				Args:        []string{"echo", "hello"},
				RedirectOut: "output.txt",
				AppendOut:   false,
			},
			expectedConsumed: 4,
		},
		{
			name:  "Command with append redirection",
			input: []string{"echo", "hello", ">>", "output.txt"},
			expectedCmd: Command{
				Args:        []string{"echo", "hello"},
				RedirectOut: "output.txt",
				AppendOut:   true,
			},
			expectedConsumed: 4,
		},
		{
			name:  "Command with input redirection",
			input: []string{"cat", "<", "input.txt"},
			expectedCmd: Command{
				Args:       []string{"cat"},
				RedirectIn: "input.txt",
			},
			expectedConsumed: 3,
		},
		{
			name:  "Command with both redirections",
			input: []string{"sort", "<", "input.txt", ">", "output.txt"},
			expectedCmd: Command{
				Args:        []string{"sort"},
				RedirectIn:  "input.txt",
				RedirectOut: "output.txt",
				AppendOut:   false,
			},
			expectedConsumed: 5,
		},
		{
			name:  "Command with append redirection after input",
			input: []string{"sort", "<", "input.txt", ">>", "output.txt"},
			expectedCmd: Command{
				Args:        []string{"sort"},
				RedirectIn:  "input.txt",
				RedirectOut: "output.txt",
				AppendOut:   true,
			},
			expectedConsumed: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCmd, consumed := parseCommandWithRedirections(tt.input)

			if consumed != tt.expectedConsumed {
				t.Errorf("Expected consumed %d, got %d", tt.expectedConsumed, consumed)
				t.Errorf("Result command: %+v", resultCmd)
				return
			}

			if len(resultCmd.Args) != len(tt.expectedCmd.Args) {
				t.Errorf("Expected %d args, got %d", len(tt.expectedCmd.Args), len(resultCmd.Args))
				t.Errorf("Expected args: %v, got: %v", tt.expectedCmd.Args, resultCmd.Args)
				return
			}

			for i, arg := range resultCmd.Args {
				if arg != tt.expectedCmd.Args[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tt.expectedCmd.Args[i], arg)
				}
			}

			if resultCmd.RedirectIn != tt.expectedCmd.RedirectIn {
				t.Errorf("RedirectIn: expected %q, got %q", tt.expectedCmd.RedirectIn, resultCmd.RedirectIn)
			}

			if resultCmd.RedirectOut != tt.expectedCmd.RedirectOut {
				t.Errorf("RedirectOut: expected %q, got %q", tt.expectedCmd.RedirectOut, resultCmd.RedirectOut)
			}

			if resultCmd.AppendOut != tt.expectedCmd.AppendOut {
				t.Errorf("AppendOut: expected %t, got %t", tt.expectedCmd.AppendOut, resultCmd.AppendOut)
			}
		})
	}
}

// Тесты для встроенных команд
func TestExecuteBuiltInEcho(t *testing.T) {
	// Перенаправляем stdout для проверки вывода
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Восстанавливаем stdout после теста
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test echo command execution
	cmd := Command{
		Args: []string{"echo", "hello", "world"},
	}

	err := executeBuiltIn(cmd)
	if err != nil {
		t.Errorf("executeBuiltIn failed: %v", err)
	}

	// Закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected output to contain 'hello world', got: %s", output)
	}
}

func TestExecuteBuiltInPwd(t *testing.T) {
	// Перенаправляем stdout для проверки вывода
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Восстанавливаем stdout после теста
	defer func() {
		os.Stdout = oldStdout
	}()

	// Test pwd command execution
	cmd := Command{
		Args: []string{"pwd"},
	}

	err := executeBuiltIn(cmd)
	if err != nil {
		t.Errorf("executeBuiltIn failed: %v", err)
	}

	// Закрываем pipe и читаем вывод
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected pwd to produce output")
	}
}

func TestExecuteBuiltInCd(t *testing.T) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Test cd command execution
	cmd := Command{
		Args: []string{"cd", ".."},
	}

	err = executeBuiltIn(cmd)
	if err != nil {
		t.Errorf("executeBuiltIn failed: %v", err)
	}

	// Change back to original directory
	cmd = Command{
		Args: []string{"cd", currentDir},
	}

	err = executeBuiltIn(cmd)
	if err != nil {
		t.Errorf("executeBuiltIn failed: %v", err)
	}
}

// Тест для команды с логическим оператором &&
func TestExecuteCommandWithAndOperator(t *testing.T) {
	// Создаем временные файлы для теста
	testFile1 := "test1.txt"
	testFile2 := "test2.txt"
	defer os.Remove(testFile1)
	defer os.Remove(testFile2)

	// Тестируем команду с оператором &&
	input := []string{"echo", "first", ">", testFile1, "&&", "echo", "second", ">", testFile2}

	// Перенаправляем stdout
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Восстанавливаем stdout после теста
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	executeCommand(input)

	// Проверяем, что оба файла созданы
	if _, err := os.Stat(testFile1); os.IsNotExist(err) {
		t.Error("First file was not created")
	}

	if _, err := os.Stat(testFile2); os.IsNotExist(err) {
		t.Error("Second file was not created")
	}
}

// Тест для команды с логическим оператором ||
func TestExecuteCommandWithOrOperator(t *testing.T) {
	// Создаем временный файл для теста
	testFile := "test_or.txt"
	defer os.Remove(testFile)

	// Тестируем команду с оператором || (первая команда заведомо успешна)
	input := []string{"echo", "success", ">", testFile, "||", "echo", "failure"}

	// Перенаправляем stdout
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Восстанавливаем stdout после теста
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	executeCommand(input)

	// Проверяем, что файл создан (первая команда выполнилась)
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File was not created, first command didn't execute")
	}
}
