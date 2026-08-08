# Ansible

Provisions a bare Ubuntu host and brings the url-shortener stack up with a single
command. Tested on Ubuntu 26.04 against three hosts in parallel.

## Requirements

**Control node**

- Ansible core 2.15 or newer
- Collections from `requirements.yml`:

```bash
ansible-galaxy collection install -r requirements.yml
```

**Target hosts**

- Ubuntu (Debian family)
- SSH access by key
- Passwordless `sudo` for the deploy user

The last two are the only manual bootstrap steps:

```bash
ssh-copy-id user@<host>
ssh -t user@<host> "echo 'user ALL=(ALL) NOPASSWD:ALL' | sudo tee /etc/sudoers.d/user && sudo chmod 440 /etc/sudoers.d/user"
```

## Layout

```
ansible.cfg                     default inventory, YAML output, pipelining, vault password file
inventory.yml                   hosts and connection settings
url_shortener_playbook.yml      entry point
group_vars/
  url_shortener/
    vars.yml                    plain variables
    vault.yml                   encrypted secrets
roles/
  common/                       base packages
  docker/                       Docker Engine from the upstream repository
  url_shortener/                application deployment
```

## Roles

| Role | Purpose |
|---|---|
| `common` | Base packages every machine should have |
| `docker` | Registers the official Docker apt repository together with its GPG key, installs Docker Engine, CLI, containerd and the Compose plugin, enables the service, adds the deploy user to the `docker` group |
| `url_shortener` | Creates the application directory, renders `.env` from a Jinja2 template, copies `docker-compose.yml`, `migrations/` and `monitoring/`, pulls images and starts the stack |

## Secrets

Secrets live in `group_vars/url_shortener/vault.yml`, encrypted with `ansible-vault`.
The encrypted file is committed; the password is not.

Create the password file outside the repository:

```bash
openssl rand -base64 32 > ~/.vault_pass && chmod 600 ~/.vault_pass
```

`ansible.cfg` already points at it, so no `--ask-vault-pass` is needed.

Working with the vault:

```bash
ansible-vault view   group_vars/url_shortener/vault.yml
ansible-vault edit   group_vars/url_shortener/vault.yml
ansible-vault rekey  group_vars/url_shortener/vault.yml
```

Variables inside the vault are prefixed with `vault_` and referenced from
`vars.yml`. That keeps variable names readable in plain text while the values
stay encrypted.

## Usage

Full run:

```bash
ansible-playbook url_shortener_playbook.yml
```

The playbook is idempotent — a second run over an unchanged host reports
`changed=0`.

### Tags

| Tag | Effect |
|---|---|
| `common` | Base packages only |
| `docker` | Docker installation only |
| `app` | Application deployment only |
| `always` | Secret validation, runs on every invocation |
| `cleanup` | Prune unused Docker objects (volumes are preserved) |
| `wipe` | **Destructive.** Remove the stack together with its data |

`cleanup` and `wipe` are also tagged `never`, so they never run as part of a
normal invocation and have to be requested explicitly:

```bash
ansible-playbook url_shortener_playbook.yml --tags app
ansible-playbook url_shortener_playbook.yml --skip-tags docker
ansible-playbook url_shortener_playbook.yml --tags cleanup
```

List what is available without running anything:

```bash
ansible-playbook url_shortener_playbook.yml --list-tags
ansible-playbook url_shortener_playbook.yml --list-tasks
```

## Variables

Defined in `roles/url_shortener/defaults/main.yml`:

| Variable | Default | Purpose |
|---|---|---|
| `app_dir` | `/opt/url_shortener` | Where the stack is deployed |
| `src_dir` | `{{ playbook_dir }}/..` | Repository root, source of the files being copied |
| `postgres_host` | `postgres` | Service name inside the Compose network |
| `postgres_port` | `5432` | |
| `postgres_user` | `postgres` | |
| `postgres_db` | `urls_and_codes` | |
| `postgres_password` | *empty* | Supplied through the vault; the playbook fails early if unset |
| `redis_host` | `redis` | |
| `redis_port` | `6379` | |

Override per host or per group through `host_vars/` and `group_vars/`.

## Linting

```bash
ansible-lint --profile production
```

The playbook passes the `production` profile. The same check runs in CI on
GitHub Actions, GitLab CI and CircleCI.

## Verify

```bash
curl http://<host>/health
```

## Notes

- Inventory group names must not contain hyphens — Ansible exposes them as Jinja
  variables, where a hyphen reads as subtraction.
- `.env` is deployed as `root:root` with mode `0600`: it holds the database
  password and only the Docker daemon needs to read it.
