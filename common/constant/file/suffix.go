package file

type Suffix string

const (
	JSON = Suffix("json")
	TOML = Suffix("toml")
	YAML = Suffix("yaml")
	YML  = Suffix("yml")
)

const (
	PlaceholderPrefix = "${"
	PlaceholderSuffix = "}"
)
