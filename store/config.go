package store

type Config struct {
	AuthyId         int    `json:"authy_id"`
	Device          Device `json:"device"`
	BackupsPassword string `json:"backups_password,omitempty"`
}

type Device struct {
	Id         int    `json:"id"`
	SecretSeed string `json:"secret_seed"`
	ApiKey     string `json:"api_key"`
}

func (c *Config) IsRegistered() bool {
	return c.AuthyId > 0 && c.Device.Id > 0 && c.Device.SecretSeed != ""
}
