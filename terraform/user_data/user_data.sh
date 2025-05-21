#!/bin/bash
apt-get update
apt-get install -y curl

sudo mkdir /tmp/ssm
cd /tmp/ssm
wget https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/debian_amd64/amazon-ssm-agent.deb
sudo dpkg -i amazon-ssm-agent.deb

systemctl enable amazon-ssm-agent
systemctl start amazon-ssm-agent

sudo apt-get install -y gcc git ffmpeg python3.11-venv

cd ~
git clone https://github.com/lastvoidtemplar/song_recognition.git

cd song_recognition

python3 -m venv venv
source venv/bin/activate
pip3 install -r requirements.txt

aws ssm get-parameter \
  --name "/audio-backend/cookies.txt" \
  --with-decryption \
  --query "Parameter.Value" \
  --output text > cookies.txt


GOLATEST=$(curl -s "https://go.dev/VERSION?m=text" | head -n 1)
GOURL=https://dl.google.com/go/${GOLATEST}.linux-amd64.tar.gz

curl -L "${GOURL}" -o /tmp/go.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
export GOCACHE=/root/.cache/go-build
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bash_profile
echo 'export GOPATH=/root/go' >> /root/.bash_profile
echo 'export GOCACHE=/root/.cache/go-build' >> /root/.bash_profile

mkdir downloads
mkdir uploads

CGO_ENABLED=1 go build -o main cmd/main.go cmd/routes.go 2> build.log
./main -region="eu-central-1" -prod > run.log
