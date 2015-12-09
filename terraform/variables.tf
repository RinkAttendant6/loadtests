variable "do_region"  { default = "tor1" }

variable "influx_dbname"   { default = "loadtests-db" }
variable "influx_username" { default = "loadtests-user"}
variable "do_ssh_key" {}

variable "do_token"   {}
variable "do_region"  { default = "nyc3" }
variable "do_ssh_key" {}

variable "influx_port"     { default = "8086" }
variable "influx_dbname"   { default = "loadtests_db" }
variable "influx_username" { default = "loadtests_user"}
variable "influx_password" {}

variable "scheduler_port"          { default = 8080 }
variable "scheduler_executor_size" { default = "512mb" }

variable "github_id"     {}
variable "github_secret" {}
