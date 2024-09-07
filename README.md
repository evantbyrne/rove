# Rove

Deploy containers over SSH.


## Architecture

Rove is intended to be a relatively simple client for managing single-server Docker Swarms, while smoothing over some of the annoyances that come with rolling your own tooling. It is designed in such a way that if you grow beyond Rove's capabilities, then self-management does not require changes to the server because there is no runtime other than Docker. Rove commands do not have unannounced side-effects to avoid interference with other aspects of server management. You will not find a privacy policy because we do not collect telemetry.

The Rove command line client connects to servers via SSH with key-based authentication. Nothing is installed by Rove on the client. When setting up a server, Rove installs Docker, enables Swarm mode, configures the firewall to allow SSH, configures the firewall to block Swarm management ports, and enables the firewall. It does not currently manage software or OS updates but may optionally in the future.


## Usage

See the [Rove homepage](https://rove.dev) for a tutorial on basic usage.

Check out the [roadmap](https://github.com/users/evantbyrne/projects/1) for planned features.

```
$ go run rove/main.go --help
Usage: main <command> [flags]

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  inspect <name> [flags]
    Inspect services and tasks.

  login <username> <password-file> [flags]
    Log into docker registries.

  logout [flags]
    Log out of docker registries.

  logs <name> [flags]
    View logs.

  machine add <name> <address> <user> <pk> [flags]

  machine delete <machine> [flags]

  machine list [flags]

  machine use <name> [flags]

  network add <name> [flags]

  network delete <name> [flags]

  network list [flags]

  secret create <name> <file> [flags]

  secret delete <name> [flags]

  secret list [flags]

  service delete <name> [flags]

  service list [flags]

  service run <name> <image> [<command> ...] [flags]

  task list [flags]

  task run <image> [<command> ...] [flags]

Run "main <command> --help" for more information on a command.
```


## Installation

Rove ships as a single binary which may be placed wherever you please, usually a directory found in the terminal `$PATH`. See the [releases section](https://github.com/evantbyrne/rove/releases) of the git repository for downloads.


## Supported Systems

Rove currently supports running the command line client on MacOS and Linux generally. It fully supports deploying to Ubuntu 24 and Debian 12 or later, with plans to more fully support other Linux distributions once firewall configuration has been implemented for systems that don't support UFW. Both x86 and Arm architectures are supported.


## Security

The `rove login` command uses `docker login` behind the scenes to authenticate with container registries. Secrets utilize Swarm's secrets storage, which mounts secrets files in the `/run/secrets` directory on configured containers. It is inadvisable to store secrets within environment variables. Rove is not designed to harden Docker installations.


## Sponsorship

Sponsorship of features that conform to Rove's values, style, and design are encouraged. Sponsorship is required for commercial integrations. An example of a feature that should be sponsored would be integrating a webhost's REST API for provisioning virtual machines via a subcommand. Commercial features must be opt-in, must not cause harm, and they must not collect telemetry. Contact [Evan](https://www.linkedin.com/in/evan-byrne-6b23a810a/) for sponsorship inquiries.
