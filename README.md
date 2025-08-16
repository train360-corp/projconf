# ProjConf

A Supabase-powered project configuration utility.

## CLI

### Getting Started

#### Linux (APT)

Install:

```shell
# Add PPA Signing Key
# Check https://train360-corp.github.io/ppa/ for update instructions, if any
curl -fsSL https://train360-corp.github.io/ppa/packages/KEY.gpg | sudo gpg --dearmor -o /usr/share/keyrings/train360-corp-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/train360-corp-keyring.gpg] https://train360-corp.github.io/ppa/packages ./" | sudo tee /etc/apt/sources.list.d/train360-corp-packages.list
sudo apt update

# Install
apt install projconf
```

Upgrade:

```shell
rm -rf /var/lib/apt/lists/train360-corp.github.io_*
apt update
apt install projconf
```

#### Homebrew (MacOS / Linux)

Install:

```shell
brew tap train360-corp/taps
brew install train360-corp/taps/projconf
```

Upgrade:

```shell
brew upgrade projconf
```