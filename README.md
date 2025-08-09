# Rove

Rove is a tool for self-hosting on Linux. It ships as a single binary which provisions machines and deploys containers directly to them via SSH.


## Quick Start

```
% rove machine add <name> <ip> <user> ~/.ssh/id_ed25519

Connected to remote address '<ip>'.

Rove will make the following changes to remote machine:

 ~ Enable firewall
 ~ Install docker
 ~ Enable swarm

Do you want Rove to run this deployment?
  Type 'yes' to approve, or anything else to deny.
  Enter a value: yes

...

Setup '<name>' and set as default machine.
```

Under the hood, Rove just setup a standard docker installation with a firewall. No proprietary or closed-source technology is ever deployed to the server. Rove does not collect telemetry. All client configuration is saved locally in a .rove file, which does not need to be synced to deploy from other workstations, CI, etc.

Next, let's deploy a python container to host a simple file server. We'll use a service for this, because we want it to run indefinitely. 

```
% rove service run files --publish 80:80 python:3.12 python3 -m http.server 80

Rove will create files:

 + service files:
 +   command  = ["python3","-m","http.server","80"]
 +   image    = "python:3.12"
 +   publish  = ["80:80"]
 +   replicas = "1"

Do you want Rove to run this deployment?
  Type 'yes' to approve, or anything else to deny.
  Enter a value: yes

...

Rove deployed 'files'.
```

Rove diffs the options you provide against what is actually running, so you can see exactly how changes will impact services before updating.

See the [Rove homepage](https://rove.dev) for the rest of the tutorial.

Check out the [roadmap](https://github.com/users/evantbyrne/projects/1) for planned features.


## Architecture

Rove is intended to be a relatively simple client for managing single-server Docker Swarms, while smoothing over some of the annoyances that come with rolling your own tooling. It is designed in such a way that if you grow beyond Rove's capabilities, then self-management does not require changes to the server because there is no runtime other than Docker. Rove commands do not have unannounced side-effects to avoid interference with other aspects of server management. You will not find a privacy policy because we do not collect telemetry.

The Rove command line client connects to servers via SSH with key-based authentication. Nothing is installed by Rove on the client. When setting up a server, Rove installs Docker, enables Swarm mode, configures the firewall to allow SSH, configures the firewall to block Swarm management ports, and enables the firewall. It does not currently manage software or OS updates but may optionally in the future.


## Installation

Rove ships as a single binary which may be placed wherever you please, usually a directory found in the terminal `$PATH`. See the [releases section](https://github.com/evantbyrne/rove/releases) of the git repository for downloads.


## Supported Systems

Rove currently supports running the command line client on MacOS and Linux generally. It fully supports deploying to Ubuntu 24 and Debian 12 or later, with plans to more fully support other Linux distributions once firewall configuration has been implemented for systems that don't support UFW. Both x86 and Arm architectures are supported.


## Managing Environments

- Run `rove machine use <name>` to switch between configured remote machines, or use the `--machine <name>` flag on individual commands.
- Deploy to your local machine by providing the `--local` flag to commands. Note that Swarm mode will need to be enabled on Docker.


## CI/CD

It is generally easiest to first setup services manually using Rove commands, and then use `rove service redeploy --force <service>` from CI/CD to pull in a new image. However, any of the builtin commands may be used in automations. Your CI/CD will need SSH access and to have run `rove machine add --force --skip <name> <ip> <user> <ssh-key>` before any other commands.

There are a number of flags available that may come in handy in an automated environment:

- `--skip` on `rove machine add` skips remote setup steps.
- `--force` skips confirmations.
- `--json` outputs JSON on success for commands that support it.


## Security

The `rove login` command uses `docker login` behind the scenes to authenticate with container registries. Secrets utilize Swarm's secrets storage, which mounts secrets files in the `/run/secrets` directory on configured containers. It is inadvisable to store secrets within environment variables. Rove is not designed to harden Docker installations.


## Sponsorship

Sponsorship of features that conform to Rove's values, style, and design are encouraged. Sponsorship is required for commercial integrations. An example of a feature that should be sponsored would be integrating a webhost's REST API for provisioning virtual machines via a subcommand. Commercial features must be opt-in, must not cause harm, and they must not collect telemetry. Contact [Evan](https://www.linkedin.com/in/evan-byrne-6b23a810a/) for sponsorship inquiries.


## Usage

```
% rove --help
Usage: rove <command> [flags]

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

  service redeploy <name> [flags]

  service rollback <name> [flags]

  service run <name> <image> [<command> ...] [flags]

  task list [flags]

  task run <image> [<command> ...] [flags]

  volume add <name> [flags]

  volume delete <name> [flags]

  volume inspect <name> [flags]

  volume list [flags]

Run "rove <command> --help" for more information on a command.
```
