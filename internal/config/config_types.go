package config

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func (c *Config) SetUser(current_user string) error {
	c.Current_user_name = current_user
	return write(*c)
}
