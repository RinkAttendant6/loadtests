provider "digitalocean" {
    token = "${var.do_token}"
}


resource "digitalocean_droplet" "influxdb" {
    name = "influxdb"
    region = "${var.do_region}"
    size = "1gb"
    image = "coreos-stable"

    ssh_keys = ["${var.do_ssh_key}"]

    connection {
        user = "core"
    }

    provisioner "file" {
        source = "files/influxdb/"
        destination = "/tmp/"
    }

    provisioner "remote-exec" {
        inline = <<CMD
sudo mkdir -p /var/opt/influxdb/
sudo mkdir -p /etc/opt/influxdb/
sudo mkdir -p /opt/influxdb/

sudo groupadd influxdb
sudo useradd influxdb -g influxdb
sudo chown -R influxdb:influxdb /var/opt/influxdb/

wget https://s3.amazonaws.com/influxdb/influxdb_0.9.4.2_x86_64.tar.gz
tar xvfz influxdb_0.9.4.2_x86_64.tar.gz
sudo mv influxdb_0.9.4.2_x86_64/opt/influxdb/versions/0.9.4.2/influxd /opt/influxdb/influxd
sudo chmod +x /opt/influxdb/influxd

sudo mv /tmp/influxdb.conf /etc/opt/influxdb/influxdb.conf
sudo mv /tmp/influxdb.service /etc/systemd/system/influxdb.service
sudo systemctl enable influxdb.service
sudo systemctl start influxdb.service
while netstat -lnt | awk '$4 ~ /:8086$/ {exit 1}'; do sleep 10; done
curl --retry 50 -G http://localhost:8086/query --data-urlencode "u=${var.influx_username}" --data-urlencode "p=${var.influx_password}" --data-urlencode "q=CREATE DATABASE ${var.influx_dbname}"
CMD
    }
}

resource "digitalocean_droplet" "scheduler" {
    name = "scheduler"
    region = "${var.do_region}"
    size = "1gb"
    image = "coreos-stable"

    ssh_keys = ["${var.do_ssh_key}"]

    connection {
        user = "core"
    }

    provisioner "local-exec" {
        command = "GOOS=linux godep go build -o ./files/scheduler/schedulerd github.com/lgpeterson/loadtests/cmd/schedulerd"
    }

    provisioner "local-exec" {
        command = "GOOS=linux godep go build -o ./files/scheduler/executord github.com/lgpeterson/loadtests/executor/cmd/executord"
    }

    provisioner "file" {
        source = "files/scheduler/"
        destination = "/tmp/"
    }

    provisioner "local-exec" {
        command = "rm ./files/scheduler/schedulerd"
    }

    provisioner "local-exec" {
        command = "rm ./files/scheduler/executord"
    }

    provisioner "remote-exec" {
        inline = <<CMD
sudo mkdir -p /opt/
sudo mv /tmp/schedulerd /opt/schedulerd
sudo mv /tmp/executord /opt/executord
sudo chmod +x /opt/schedulerd
sudo mkdir -p /etc/scheduler/

cat >> /tmp/scheduler.env <<EOF
INFLUX_ADDR=${digitalocean_droplet.influxdb.ipv4_address}:${var.influx_port}
INFLUX_DB_NAME=${var.influx_dbname}
INFLUX_USERNAME=${var.influx_username}
INFLUX_PASSWORD=${var.influx_password}
DO_TOKEN=${var.do_token}
EXECUTOR_BINARY_FILEPATH=/opt/executord

PORT=${var.scheduler_port}
DROPLET_REGION=${var.do_region}
DROPLET_SIZE=${var.scheduler_executor_size}
EOF

sudo mv /tmp/scheduler.env /etc/scheduler/scheduler.env
sudo mv /tmp/scheduler.service /etc/systemd/system/scheduler.service
sudo systemctl enable scheduler.service
sudo systemctl start scheduler.service
CMD
    }
}

