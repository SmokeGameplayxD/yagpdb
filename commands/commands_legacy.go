package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"

	"github.com/mediocregopher/radix.v2/redis"
)

type legacyChannelCommandSetting struct {
	Info           *YAGCommand `json:"-"` // Used for template info
	Cmd            string      `json:"cmd"`
	CommandEnabled bool        `json:"enabled"`
	AutoDelete     bool        `json:"autodelete"`
	RequiredRole   string      `json:"required_role"`
}

type legacyChannelOverride struct {
	Settings        []*legacyChannelCommandSetting `json:"settings"`
	OverrideEnabled bool                           `json:"enabled"`
	Channel         string                         `json:"channel"`
	ChannelName     string                         `json:"-"` // Used for the template rendering
}

type legacyCommandsConfig struct {
	Prefix string `json:"-"` // Stored in a seperate key for speed

	Global           []*legacyChannelCommandSetting `json:"gloabl"`
	ChannelOverrides []*legacyChannelOverride       `json:"overrides"`
}

// Fills in the defaults for missing data, for when users create channels or commands are added
func legacyCheckChannelsConfig(conf *legacyCommandsConfig, channels []*discordgo.Channel) {

	commands := CommandSystem.Root.Commands

	if conf.Global == nil {
		conf.Global = []*legacyChannelCommandSetting{}
	}

	if conf.ChannelOverrides == nil {
		conf.ChannelOverrides = []*legacyChannelOverride{}
	}

ROOT:
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		// Look for an existing override
		for _, override := range conf.ChannelOverrides {
			// Found an existing override, check if it has all the commands
			if channel.ID == override.Channel {
				override.Settings = legacycheckCommandSettings(override.Settings, commands, false)
				override.ChannelName = channel.Name // Update name if changed
				continue ROOT
			}
		}

		// Not found, create a default override
		override := &legacyChannelOverride{
			Settings:        []*legacyChannelCommandSetting{},
			OverrideEnabled: false,
			Channel:         channel.ID,
			ChannelName:     channel.Name,
		}

		// Fill in default command settings
		override.Settings = legacycheckCommandSettings(override.Settings, commands, false)
		conf.ChannelOverrides = append(conf.ChannelOverrides, override)
	}

	newOverrides := make([]*legacyChannelOverride, 0, len(conf.ChannelOverrides))

	// Check for removed channels
	for _, override := range conf.ChannelOverrides {
		for _, channel := range channels {
			if channel.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			if channel.ID == override.Channel {
				newOverrides = append(newOverrides, override)
				break
			}
		}
	}
	conf.ChannelOverrides = newOverrides

	// Check the global settings
	conf.Global = legacycheckCommandSettings(conf.Global, commands, true)
}

// Checks a single list of ChannelCommandSettings and applies defaults if not found
func legacycheckCommandSettings(settings []*legacyChannelCommandSetting, commands []*dcmd.RegisteredCommand, defaultEnabled bool) []*legacyChannelCommandSetting {

ROOT:
	for _, registeredCmd := range commands {
		cast, ok := registeredCmd.Command.(*YAGCommand)
		if !ok {
			continue
		}

		for _, settingsCmd := range settings {
			if cast.Name == settingsCmd.Cmd {
				// Bingo
				settingsCmd.Info = cast
				continue ROOT
			}
		}

		// Not found, add it to the list of overrides
		settingsCmd := &legacyChannelCommandSetting{
			Cmd:            cast.Name,
			CommandEnabled: defaultEnabled,
			AutoDelete:     false,
			Info:           cast,
		}
		settings = append(settings, settingsCmd)
	}

	newSettings := make([]*legacyChannelCommandSetting, 0, len(settings))

	// Check for commands that have been removed (e.g the config contains commands from an older version)
	for _, settingsCmd := range settings {
		for _, registeredCmd := range commands {
			cast, ok := registeredCmd.Command.(*YAGCommand)
			if !ok {
				continue
			}

			if cast.Name == settingsCmd.Cmd {
				newSettings = append(newSettings, settingsCmd)
				break
			}
		}
	}

	return newSettings
}

func legacyGetConfig(client *redis.Client, guild string, channels []*discordgo.Channel) *legacyCommandsConfig {
	var config *legacyCommandsConfig
	err := common.GetRedisJson(client, "commands_settings:"+guild, &config)
	if err != nil {
		log.WithError(err).Error("Error retrieving command settings")
	}

	if config == nil {
		config = &legacyCommandsConfig{}
	}

	prefix, err := GetCommandPrefix(client, guild)
	if err != nil {
		// Continue as normal with defaults
		log.WithError(err).Error("Error fetching command prefix")
	}

	config.Prefix = prefix

	// Fill in defaults
	legacyCheckChannelsConfig(config, channels)

	return config
}
