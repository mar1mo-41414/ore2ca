package cmd

import (
	"fmt"
	"strings"

	"github.com/ore2ca/ore2ca/internal/store"
	"github.com/spf13/cobra"
)

func newDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker <target>",
		Short: "Docker向け設定例を生成する",
		Long: `各種リバースプロキシのDocker Compose設定例を生成します。

利用可能なターゲット: nginx, caddy, traefik`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"nginx", "caddy", "traefik"},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := strings.ToLower(args[0])
			s, err := store.New()
			if err != nil {
				return err
			}
			switch target {
			case "nginx":
				return printNginxConfig(s)
			case "caddy":
				return printCaddyConfig(s)
			case "traefik":
				return printTraefikConfig(s)
			default:
				return fmt.Errorf("不明なターゲット: %s (nginx, caddy, traefik のいずれかを指定)", target)
			}
		},
	}
	return cmd
}

func printNginxConfig(s *store.Store) error {
	fmt.Printf(`# nginx + ore2ca 設定例
# ore2ca issue myapp.local を実行後に使用してください

# docker-compose.yml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - %s:/etc/ssl/certs/ore2ca.crt:ro
      - %s:/etc/ssl/private/myapp.key:ro
      - %s:/etc/ssl/private/myapp.crt:ro

# nginx.conf
server {
    listen 443 ssl;
    server_name myapp.local;

    ssl_certificate     /etc/ssl/private/myapp.crt;
    ssl_certificate_key /etc/ssl/private/myapp.key;

    location / {
        proxy_pass http://app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name myapp.local;
    return 301 https://$host$request_uri;
}
`,
		s.CACertPath(),
		s.KeyPath("myapp.local"),
		s.CertPath("myapp.local"),
	)
	return nil
}

func printCaddyConfig(s *store.Store) error {
	fmt.Printf(`# Caddy + ore2ca 設定例
# ore2ca issue myapp.local を実行後に使用してください

# docker-compose.yml
services:
  caddy:
    image: caddy:alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - %s:/etc/caddy/ore2ca.crt:ro
      - %s:/etc/caddy/myapp.crt:ro
      - %s:/etc/caddy/myapp.key:ro

# Caddyfile
myapp.local {
    tls /etc/caddy/myapp.crt /etc/caddy/myapp.key
    reverse_proxy app:8080
}
`,
		s.CACertPath(),
		s.CertPath("myapp.local"),
		s.KeyPath("myapp.local"),
	)
	return nil
}

func printTraefikConfig(s *store.Store) error {
	fmt.Printf(`# Traefik + ore2ca 設定例
# ore2ca issue myapp.local を実行後に使用してください

# docker-compose.yml
services:
  traefik:
    image: traefik:v3.0
    ports:
      - "443:443"
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./certs.yml:/etc/traefik/dynamic/certs.yml:ro
      - %s:/certs/myapp.crt:ro
      - %s:/certs/myapp.key:ro

  app:
    image: your-app
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.app.rule=Host(` + "`" + `myapp.local` + "`" + `)"
      - "traefik.http.routers.app.tls=true"
      - "traefik.http.routers.app.entrypoints=websecure"

# traefik.yml
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    exposedByDefault: false
  file:
    directory: /etc/traefik/dynamic

# certs.yml (dynamic config)
tls:
  certificates:
    - certFile: /certs/myapp.crt
      keyFile: /certs/myapp.key
`,
		s.CertPath("myapp.local"),
		s.KeyPath("myapp.local"),
	)
	return nil
}
