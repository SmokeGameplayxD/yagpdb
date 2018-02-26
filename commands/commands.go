package commands

//go:generate esc -o assets_gen.go -pkg commands -ignore ".go" assets/
//go:generate sqlboiler --no-hooks -w "commands_channels_overrides,commands_channels_cmd_overrides" postgres

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jonas747/yagpdb/commands/models"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/docs"
	"github.com/mediocregopher/radix.v2/redis"
)

type Plugin struct{}

func RegisterPlugin() {
	plugin := &Plugin{}
	common.RegisterPlugin(plugin)
	err := common.GORM.AutoMigrate(&common.LoggedExecutedCommand{}).Error
	if err != nil {
		log.WithError(err).Error("Failed migrating database")
	}

	docs.AddPage("Commands", FSMustString(false, "/assets/help-page.md"), nil)
}

func (p *Plugin) Name() string {
	return "Commands"
}

func GetCommandPrefix(client *redis.Client, guild string) (string, error) {
	reply := client.Cmd("GET", "command_prefix:"+guild)
	if reply.Err != nil {
		return "", reply.Err
	}
	if reply.IsType(redis.Nil) {
		return "", nil
	}

	return reply.Str()
}

// Fills in defaults for non existing commands
func GetFullCommandsSettings(settings *models.CommandsChannelsOverride) {
}
