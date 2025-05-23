package cmd

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoibubbles"
)

// readBool reads a bool.
func (c *Config) readBool(prompt string, defaultValue *bool) (bool, error) {
	switch {
	case c.noTTY:
		fullPrompt := prompt
		if defaultValue != nil {
			fullPrompt += " (default " + strconv.FormatBool(*defaultValue) + ")"
		}
		fullPrompt += "? "
		for {
			valueStr, err := c.readLineRaw(fullPrompt)
			if err != nil {
				return false, err
			}
			if valueStr == "" && defaultValue != nil {
				return *defaultValue, nil
			}
			if value, err := chezmoi.ParseBool(valueStr); err == nil {
				return value, nil
			}
		}
	default:
		initModel := chezmoibubbles.NewBoolInputModel(prompt, defaultValue)
		finalModel, err := runCancelableModel(initModel)
		if err != nil {
			return false, err
		}
		return finalModel.Value(), nil
	}
}

// readChoice reads a choice.
func (c *Config) readChoice(prompt string, choices []string, defaultValue *string) (string, error) {
	switch {
	case c.noTTY:
		fullPrompt := prompt + " (" + strings.Join(choices, "/")
		if defaultValue != nil {
			fullPrompt += ", default " + *defaultValue
		}
		fullPrompt += ")? "
		abbreviations := chezmoi.UniqueAbbreviations(choices)
		for {
			value, err := c.readLineRaw(fullPrompt)
			if err != nil {
				return "", err
			}
			if value == "" && defaultValue != nil {
				return *defaultValue, nil
			}
			if value, ok := abbreviations[value]; ok {
				return value, nil
			}
		}
	default:
		initModel := chezmoibubbles.NewChoiceInputModel(prompt, choices, defaultValue)
		finalModel, err := runCancelableModel(initModel)
		if err != nil {
			return "", err
		}
		return finalModel.Value(), nil
	}
}

// readInt reads an int.
func (c *Config) readInt(prompt string, defaultValue *int64) (int64, error) {
	switch {
	case c.noTTY:
		fullPrompt := prompt
		if defaultValue != nil {
			fullPrompt += " (default " + strconv.FormatInt(*defaultValue, 10) + ")"
		}
		fullPrompt += "? "
		for {
			valueStr, err := c.readLineRaw(fullPrompt)
			if err != nil {
				return 0, err
			}
			if valueStr == "" && defaultValue != nil {
				return *defaultValue, nil
			}
			if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
				return value, nil
			}
		}
	default:
		initModel := chezmoibubbles.NewIntInputModel(prompt, defaultValue)
		finalModel, err := runCancelableModel(initModel)
		if err != nil {
			return 0, err
		}
		return finalModel.Value(), nil
	}
}

// readLineRaw reads a line, trimming leading and trailing whitespace.
func (c *Config) readLineRaw(prompt string) (string, error) {
	_, err := c.stdout.Write([]byte(prompt))
	if err != nil {
		return "", err
	}
	if c.bufioReader == nil {
		c.bufioReader = bufio.NewReader(c.stdin)
	}
	line, err := c.bufioReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// readMultichoice reads multiple choices from a list.
func (c *Config) readMultichoice(prompt string, choices []string, defaultValue *[]string) ([]string, error) {
	switch {
	case c.noTTY:
		shortPrompt := "Choice (ENTER to stop)> "

		fullPrompt := prompt + "?\nChoices (" + strings.Join(choices, "/") + ")"
		if defaultValue != nil {
			fullPrompt += "\nDefault [" + strings.Join(*defaultValue, ", ") + "]"
		}
		fullPrompt += "\n"

		_, err := c.stdout.Write([]byte(fullPrompt))
		if err != nil {
			return []string{}, err
		}

		abbreviations := chezmoi.UniqueAbbreviations(choices)
		selected := make(map[string]struct{})

		for {
			value, err := c.readLineRaw(shortPrompt)
			if err != nil {
				return []string{}, err
			}

			if value == "" {
				if len(selected) == 0 && defaultValue != nil {
					return *defaultValue, nil
				}

				if len(selected) > 0 {
					break
				}
			}

			if value == "[]" {
				return []string{}, nil
			}

			if value, ok := abbreviations[value]; ok {
				selected[value] = struct{}{}
			}

			if len(selected) == len(choices) {
				break
			}
		}

		result := make([]string, 0)
		for name := range selected {
			result = append(result, name)
		}

		return result, nil

	default:
		initModel := chezmoibubbles.NewMultichoiceInputModel(prompt, choices, defaultValue)
		finalModel, err := tea.NewProgram(initModel, tea.WithOutput(os.Stderr)).Run()
		if err != nil {
			return []string{}, err
		}

		return finalModel.(chezmoibubbles.MultichoiceInputModel).Value(), nil //nolint:forcetypeassert
	}
}

// readPassword reads a password.
func (c *Config) readPassword(prompt, placeholder string) (string, error) {
	switch {
	case c.noTTY:
		return c.readLineRaw(prompt)
	case c.PINEntry.Command != "":
		return c.readPINEntry(prompt)
	default:
		initModel := chezmoibubbles.NewPasswordInputModel(prompt, placeholder)
		finalModel, err := runCancelableModel(initModel)
		if err != nil {
			return "", err
		}
		return finalModel.Value(), nil
	}
}

// readString reads a string.
func (c *Config) readString(prompt string, defaultValue *string) (string, error) {
	switch {
	case c.noTTY:
		fullPrompt := prompt
		if defaultValue != nil {
			fullPrompt += " (default " + strconv.Quote(*defaultValue) + ")"
		}
		fullPrompt += "? "
		value, err := c.readLineRaw(fullPrompt)
		if err != nil {
			return "", err
		}
		if value == "" && defaultValue != nil {
			return *defaultValue, nil
		}
		return value, nil
	default:
		initModel := chezmoibubbles.NewStringInputModel(prompt, defaultValue)
		finalModel, err := runCancelableModel(initModel)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(finalModel.Value()), nil
	}
}

func (c *Config) promptBool(prompt string, args ...bool) (bool, error) {
	var defaultValue *bool
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		defaultValue = &args[0]
	default:
		return false, fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
	}
	if c.interactiveTemplateFuncs.promptDefaults && defaultValue != nil {
		return *defaultValue, nil
	}
	return c.readBool(prompt, defaultValue)
}

// promptChoice prompts the user for one of choices until a valid choice is made.
func (c *Config) promptChoice(prompt string, choices []string, args ...string) (string, error) {
	var defaultValue *string
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:

		if !slices.Contains(choices, args[0]) {
			return "", fmt.Errorf("%s: invalid default value", args[0])
		}
		defaultValue = &args[0]
	default:
		return "", fmt.Errorf("want 2 or 3 arguments, got %d", len(args)+2)
	}
	if c.interactiveTemplateFuncs.promptDefaults && defaultValue != nil {
		return *defaultValue, nil
	}
	return c.readChoice(prompt, choices, defaultValue)
}

func (c *Config) promptInt(prompt string, args ...int64) (int64, error) {
	var defaultValue *int64
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		defaultValue = &args[0]
	default:
		return 0, fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
	}
	if c.interactiveTemplateFuncs.promptDefaults && defaultValue != nil {
		return *defaultValue, nil
	}
	return c.readInt(prompt, defaultValue)
}

// promptMultiChoice prompts the user to select one or more values in a list.
func (c *Config) promptMultichoice(prompt string, choices []string, defaults *[]string) ([]string, error) {
	if defaults != nil {
		for i, defaultValue := range *defaults {
			if !slices.Contains(choices, defaultValue) {
				return []string{}, fmt.Errorf("%s: invalid default value (index %d)", defaultValue, i+1)
			}
		}
	}

	if c.interactiveTemplateFuncs.promptDefaults && defaults != nil {
		return *defaults, nil
	}

	return c.readMultichoice(prompt, choices, defaults)
}

func (c *Config) promptString(prompt string, args ...string) (string, error) {
	var defaultValue *string
	switch len(args) {
	case 0:
		// Do nothing.
	case 1:
		arg := strings.TrimSpace(args[0])
		defaultValue = &arg
	default:
		return "", fmt.Errorf("want 1 or 2 arguments, got %d", len(args)+1)
	}
	if c.interactiveTemplateFuncs.promptDefaults && defaultValue != nil {
		return *defaultValue, nil
	}
	return c.readString(prompt, defaultValue)
}

type cancelableModel interface {
	tea.Model
	Canceled() bool
}

func runCancelableModel[M cancelableModel](initModel M) (M, error) {
	switch finalModel, err := runModel(initModel); {
	case err != nil:
		return finalModel, err
	case finalModel.Canceled():
		return finalModel, chezmoi.ExitCodeError(0)
	default:
		return finalModel, nil
	}
}

func runModel[M tea.Model](initModel M) (M, error) {
	program := tea.NewProgram(initModel)
	finalModel, err := program.Run()
	return finalModel.(M), err //nolint:forcetypeassert,revive
}
