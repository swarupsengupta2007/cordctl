package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type OptionType uint8

const (
	OptionTypeString  = OptionType(discordgo.ApplicationCommandOptionString)
	OptionTypeBoolean = OptionType(discordgo.ApplicationCommandOptionBoolean)
)

type Option struct {
	Name        string
	Description string
	Required    bool
	Type        OptionType
}

type Command struct {
	Description string
	Options     []Option
	Callback    func(cmd string, options map[string]any) string
}

type Bot struct {
	Token    string
	Commands map[string]Command
	guildID  string
	session  *discordgo.Session
}

func (b *Bot) Run() {

	if b.Token == "" {
		fmt.Println("Bot token is not set. Please provide a valid token.")
		return
	}
	if len(b.Commands) == 0 {
		fmt.Println("No commands registered. Please add commands to the bot.")
		return
	}
	fmt.Println("Starting Discord bot...")

	var err error
	b.session, err = discordgo.New("Bot " + b.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	fmt.Println("Connecting to Discord...")
	err = b.session.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Connected to Discord.")
	defer b.session.Close()

	ok := b.getGuilds()
	if !ok {
		fmt.Println("Failed to fetch guilds. Ensure the bot has the correct permissions.")
		return
	}

	commands, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		fmt.Println("error fetching commands,", err)
		return
	}

	for _, cmd := range commands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID)
		if err != nil {
			fmt.Printf("failed to delete command %s: %v\n", cmd.Name, err)
		} else {
			fmt.Printf("deleted command: %s\n", cmd.Name)
		}
	}

	ok = b.registerCommands()
	if !ok {
		fmt.Println("Failed to register commands.")
		return
	}
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.interactionHandler(s, i)
	})

	fmt.Println("Bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}

func (b *Bot) getGuilds() bool {
	guilds, err := b.session.UserGuilds(1, "", "", false)
	if err != nil {
		fmt.Println("error fetching guilds,", err)
		return false
	}

	fmt.Println("Connected to Discord as:", b.session.State.User.Username)
	fmt.Println("Guilds available:")
	for _, guild := range guilds {
		fmt.Printf("- %s (ID: %s)\n", guild.Name, guild.ID)
	}

	b.guildID = guilds[0].ID
	return true
}

func (b *Bot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	fmt.Printf("Received interaction: %s, Command: %s\n", i.Interaction.ID, data.Name)

	cmd, ok := b.Commands[data.Name]
	if !ok {
		fmt.Printf("Unknown command: %s\n", data.Name)
		sendResponse(s, i, "Unknown command. Please try again.")
		return
	}

	optval := make(map[string]any)
	for _, opt := range data.Options {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionString:
			optval[opt.Name] = opt.StringValue()
		case discordgo.ApplicationCommandOptionBoolean:
			optval[opt.Name] = opt.BoolValue()
		}
	}
	fmt.Printf("Command options: %v\n", optval)
	if cmd.Callback != nil {
		response := cmd.Callback(data.Name, optval)
		if response != "" {
			sendResponse(s, i, response)
			return
		}
	}
}

func sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	fmt.Println("Responding to interaction:", i.Interaction.ID, "with message:", message)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		fmt.Println("Failed to respond to interaction:", err)
	}
}

func (b *Bot) registerCommands() bool {
	commands := make([]*discordgo.ApplicationCommand, 0, len(b.Commands))
	for name, cmd := range b.Commands {
		options := make([]*discordgo.ApplicationCommandOption, 0, len(cmd.Options))
		for _, opt := range cmd.Options {
			options = append(options, &discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionType(opt.Type),
				Name:        opt.Name,
				Description: opt.Description,
				Required:    opt.Required,
			})
		}

		commands = append(commands, &discordgo.ApplicationCommand{
			Name:        name,
			Description: cmd.Description,
			Options:     options,
			GuildID:     b.guildID,
		})
	}
	error := false
	for _, cmd := range commands {
		fmt.Printf("Registering command: %s\n", cmd.Name)
		command, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
		if err != nil {
			fmt.Println("Cannot create command:", cmd.Name, err)
			error = true
			break
		}
		fmt.Printf("Registered command: %s, ID: %s\n", command.Name, command.ID)
	}
	if error {
		for _, cmd := range commands {
			_ = b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID)
		}
	}
	return !error
}
