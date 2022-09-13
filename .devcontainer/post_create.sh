#!/bin/bash

sudo apt-get update
sudo apt-get install -y curl postgresql-client ripgrep

wget https://github.com/ahmetb/kubectx/releases/download/v0.9.4/kubectx
sudo mv kubectx /usr/local/bin/kubectx
chmod +x /usr/local/bin/kubectx
wget https://github.com/ahmetb/kubectx/releases/download/v0.9.4/kubens
sudo mv kubens /usr/local/bin/kubens
chmod +x /usr/local/bin/kubens
