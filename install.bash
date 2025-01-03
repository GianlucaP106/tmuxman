#! /bin/bash

echo "Installing tmuxman..."

echo "Initializing configuration directory at ~/.tmuxman"
mkdir ~/.tmuxman/ 2>/dev/null | true
cd ~/.tmuxman

latest_release=$(curl -sL "https://api.github.com/repos/GianlucaP106/tmuxman/releases/latest")
latest_tag=$(echo "$latest_release" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

echo "Installing latest release ${latest_tag}"

os="$(uname -s)"
arch="$(uname -m)"

if [ "$arch" = "aarch64" ]; then
    arch="arm64"
fi

echo "Detected platform ${os}-${arch}"

file="tmuxman_${os}_${arch}.tar.gz"

curl -s -L -o build.tar.gz https://github.com/GianlucaP106/tmuxman/releases/download/${latest_tag}/${file}

tar -xzf build.tar.gz 2>/dev/null
if [ $? -ne 0 ]; then
    echo "Failed to install tmuxman. Platform not supported."
    exit
fi

rm build.tar.gz

directory="$HOME/.tmuxman"
if echo "$PATH" | grep -q "$directory"; then
    echo "Successfully installed tmuxman at ~/.tmuxman!"
else
    echo "Adding ~/.tmuxman to PATH"
    rc_files=(~/.bashrc ~/.zshrc ~/.profile ~/.bash_profile ~/.bash_login ~/.cshrc ~/.tcshrc)
    for file in "${rc_files[@]}"; do
        if [ -f "$file" ]; then
            echo "Adding to $file"
            echo '# tmuxman path export' >>"$file"
            echo 'export PATH="$PATH:$HOME/.tmuxman"' >>"$file"
            break
        fi
    done
    echo "Successfully installed tmuxman at ~/.tmuxman!"
    echo "Restart your terminal session."
fi
