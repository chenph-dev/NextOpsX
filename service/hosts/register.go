package hosts

// Register the connectors shown in the asset form: built-in command drivers
// (ssh/winrm) plus the database and middleware drivers.
import (
	_ "nextopsx/domain/es"
	_ "nextopsx/domain/redis"
	_ "nextopsx/domain/register"
)
