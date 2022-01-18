package api

import (
	"fmt"
	"strings"
)

func (s Server) DatabaseDumpCmd(d Database) ([]string, error) {
	cmd := make([]string, 0, 32)
	cmd = append(cmd, "mysqldump", "--single-transaction")

	if d.ExcludeTables != "" {
		for _, table := range strings.Fields(d.ExcludeTables) {
			cmd = append(cmd, "--ignore-table", d.Name+"."+table)
		}
	}

	cmd = append(cmd, d.Name)

	if d.OnlyTables != "" {
		cmd = append(cmd, strings.Fields(d.OnlyTables)...)
	}

	return s.wrapCmd(cmd)
}

func (s Server) DatabaseListCmd() ([]string, error) {
	ignoreTables := []string{
		"innodb",
		"mysql",
		"information_schema",
		"performance_schema",
		"sys",
		"tmp",
	}
	for i := range ignoreTables {
		ignoreTables[i] = "'" + ignoreTables[i] + "'"
	}

	cmd := []string{
		"mysql",
		"--skip-column-names",
		"--batch",
		"--execute",
		fmt.Sprintf(
			"SHOW DATABASES WHERE `Database` NOT IN(%s)",
			strings.Join(ignoreTables, ", "),
		),
	}
	return s.wrapCmd(cmd)
}

func (s Server) addAuth(cmd []string) ([]string, error) {
	parts := make([]string, 0, len(cmd)+8) // 8 is arbitrary, could be 5
	parts = append(parts, cmd[0], "--host", s.Host, "--port", fmt.Sprint(s.Port), "--user", s.Username)
	if s.Password != "" {
		password, err := s.DecryptPassword()
		if err != nil {
			return nil, err
		}
		parts = append(parts, fmt.Sprintf("--password=%s", password))
	}
	parts = append(parts, cmd[1:]...)
	return parts, nil
}

func (s Server) addProxy(cmd []string) ([]string, error) {
	if s.ProxyHost == "" {
		return cmd, nil
	}

	userHost := fmt.Sprintf("%s@%s", s.ProxyUsername, s.ProxyHost)
	proxy := []string{"ssh", "-i", s.ProxyIdentity, userHost}

	parts := make([]string, 0, len(proxy)+len(cmd))
	parts = append(parts, proxy...)
	parts = append(parts, cmd...)
	return parts, nil
}

func (s Server) wrapCmd(cmd []string) ([]string, error) {
	cmd, err := s.addAuth(cmd)
	if err != nil {
		return nil, err
	}

	cmd, err = s.addProxy(cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
