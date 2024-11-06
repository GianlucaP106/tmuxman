# tmuxman

A TUI to interact with tmux. Provides session, window and pane management with a preview of the client.

![tmuxman](https://github.com/user-attachments/assets/113dc3b6-8f50-4f26-b107-93240732331a)

## Installation

### Try with docker first

```bash
docker run -it --name tmuxman --rm ubuntu bash -c '
        apt update &&
        apt install -y git golang-go tmux curl unzip &&
        cd &&
        (curl -fsSL https://raw.githubusercontent.com/GianlucaP106/tmuxman/main/install.bash | bash) &&
        export PATH="$PATH:$HOME/.tmuxman" &&
        tmuxman
    '
```

### Binary installation

...

### Build from source

```bash
mkdir ~/.tmuxman/ 2>/dev/null | true
cd ~/.tmuxman

# clone the repo
git clone https://github.com/GianlucaP106/tmuxman src
cd src

# build the project
go build -o ~/.tmuxman/tmuxman

# add to path
export PATH="$PATH:$HOME/.tmuxman"

# optionally delete the source code
rm -rf ~/.tmuxman/src
```

>Note: tmux is required on the system. If it is not installed, tmuxman will attempt to install it.

#### Supported paltforms

Tmuxman is only supported for Linux and MacOS

## Usage

```bash
tmuxman
```

The above command will open a TUI (text based user-interface) which shows a view of sessions, windows and panes. The TUI allows to create new sessions, rename sessions and windows, kill sessions, windows and panes and more!

## Features

- Tree view of sessions, windows and panes.
- Table view of sessions, windows and panes.
- Create, update, kill sessions, windows and panes.

## Help

Use the '?' key in the TUI to see the cheatsheet
