CREATE TABLE IF NOT EXISTS commands_channels_overrides (
	id bigserial PRIMARY KEY,
	guild_id bigint NOT NULL,
	channel_ids bigint[] NOT NULL,
	is_global bool NOT NULL,

	override_enabled bool NOT NULL,
	hide_cmd_disabled_msg bool NOT NULL,

	req_role_new_cmds bigint[] NOT NULL,
	disable_new_cmds bool NOT NULL,
	autodelete_new_cmds bool NOT NULL
);

CREATE INDEX IF NOT EXISTS commands_channels_overrides_gidx ON commands_channels_overrides(guild_id);

CREATE TABLE IF NOT EXISTS commands_channels_cmd_overrides (
	id bigserial PRIMARY KEY,
	channels_overrides_id bigint references commands_channels_overrides(id) NOT NULL,

	enabled bool NOT NULL,
	auto_delete bool NOT NULL,
	required_roles bigint[] NOT NULL
);

CREATE INDEX IF NOT EXISTS commands_channels_cmd_overrides_channel_overrides_idx ON commands_channels_cmd_overrides(channels_overrides_id);