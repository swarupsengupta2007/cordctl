package bot

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type OptionTypeToken uint8

const (
	OptionTypeMandatoryValue OptionTypeToken = iota
	OptionTypeOptionalValue
	OptionTypeOptionalToken
	OptionTypeOptionalPresence
	OptionTypeAsIs
)

type argsContainer struct {
	Pre   string
	Token string
	Post  string
	Type  OptionTypeToken
}

type yamlRoot struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Command     string          `yaml:"command"`
	Args        []string        `yaml:"args"`
	Container   []argsContainer `yaml:"-"`
}

func ParseYAMLCommand(path string) (string, Command, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", Command{}, fmt.Errorf("failed to open yaml file: %w", err)
	}
	defer file.Close()

	var root yamlRoot
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		return "", Command{}, fmt.Errorf("failed to decode yaml: %w", err)
	}

	container := make([]argsContainer, len(root.Args))
	for idx, arg := range root.Args {
		if strings.Contains(arg, "{") && strings.Contains(arg, "}") {
			start := strings.Index(arg, "{")
			end := strings.Index(arg, "}")
			token := arg[start+1 : end]
			pre := arg[:start]
			post := arg[end+1:]
			if len(token) > 0 {
				switch token[0] {
				case '$':
					container[idx] = argsContainer{Pre: pre, Token: token[1:], Post: post, Type: OptionTypeMandatoryValue}
				case '?':
					container[idx] = argsContainer{Pre: pre, Token: token[1:], Post: post, Type: OptionTypeOptionalValue}
				case '=':
					container[idx] = argsContainer{Pre: pre, Token: token[1:], Post: post, Type: OptionTypeOptionalToken}
				case '!':
					container[idx] = argsContainer{Pre: pre, Token: token[1:], Post: post, Type: OptionTypeOptionalPresence}
				default:
					container[idx] = argsContainer{Pre: pre, Token: "{" + token + "}", Post: post, Type: OptionTypeAsIs}
				}
			} else {
				container[idx] = argsContainer{Pre: pre, Token: "{" + token + "}", Post: post, Type: OptionTypeAsIs}
			}
		} else {
			container[idx] = argsContainer{Pre: arg, Token: "", Post: "", Type: OptionTypeAsIs}
		}
	}
	root.Container = container
	// Build opts from container
	optsMap := make(map[string]Option)
	for _, ac := range container {
		if ac.Type == OptionTypeAsIs || ac.Token == "" {
			continue
		}
		if _, exists := optsMap[ac.Token]; exists {
			continue
		}
		optType := OptionTypeString
		if ac.Type == OptionTypeOptionalPresence || ac.Type == OptionTypeOptionalToken {
			optType = OptionTypeBoolean
		}
		optsMap[ac.Token] = Option{
			Name:        ac.Token,
			Type:        optType,
			Description: ac.Token,
			Required:    ac.Type == OptionTypeMandatoryValue,
		}
	}
	// Build optsSlice with required options first, then optional
	optsSlice := make([]Option, 0, len(optsMap))
	// First, append required options
	for _, o := range optsMap {
		if o.Required {
			optsSlice = append(optsSlice, o)
		}
	}
	// Then, append optional options
	for _, o := range optsMap {
		if !o.Required {
			optsSlice = append(optsSlice, o)
		}
	}
	cmd := Command{
		Description: root.Description,
		Options:     optsSlice,
		Callback: func(cmd string, options map[string]any) string {
			args := []string{}
			for _, acVal := range root.Container {
				switch acVal.Type {
				case OptionTypeMandatoryValue:
					if val, ok := options[acVal.Token]; ok && val != nil && val != "" {
						valStr := fmt.Sprintf("%v", val)
						args = append(args, strings.TrimSpace(acVal.Pre+valStr+acVal.Post))
					}
				case OptionTypeOptionalValue:
					if val, ok := options[acVal.Token]; ok && val != nil && val != "" {
						valStr := fmt.Sprintf("%v", val)
						args = append(args, strings.TrimSpace(acVal.Pre+valStr+acVal.Post))
					}
				case OptionTypeOptionalToken:
					if _, ok := options[acVal.Token]; ok {
						args = append(args, strings.TrimSpace(acVal.Pre+acVal.Token+acVal.Post))
					}
				case OptionTypeOptionalPresence:
					if _, ok := options[acVal.Token]; ok && options[acVal.Token] != nil && options[acVal.Token].(bool) {
						args = append(args, strings.TrimSpace(acVal.Pre+acVal.Post))
					}
				case OptionTypeAsIs:
					args = append(args, strings.TrimSpace(acVal.Pre+acVal.Token+acVal.Post))
				}
			}
			cmdline := append([]string{root.Command}, args...)
			output, err := runCommand(cmdline)
			if err != nil {
				return "Error executing command: " + err.Error()
			}
			return output
		},
	}
	return root.Name, cmd, nil
}

func runCommand(cmdline []string) (string, error) {
	fmt.Println("Running command:", cmdline)
	if len(cmdline) == 0 {
		return "", fmt.Errorf("no command to run")
	}
	out, err := exec.Command(cmdline[0], cmdline[1:]...).CombinedOutput()
	return string(out), err
}
