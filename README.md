# loadtests.me

All your load tests are belong to us

## Requirements

Before you start, you will need to install [Terraform](https://terraform.io/).

Terraform is a tool used to build, manage, and version infrastructure.

## Installation and deployment

 1. Download and unzip these files. Alternatively, clone this repository onto
    your hard disk.
 2. In this directory, `cd terraform`.
 3. Run `terraform apply`. Terraform will ask you to fill in some required
    variables such as your GitHub application secret.

These variables, along with other variables found in `/terraform/variables.tf`,
can be configured using environment variables on your system. Refer to the
Terraform documentation for more information.

At any time, you can run `terraform show` inside the `/terraform` directory to
view the status of the deployed infrastructure.

## Uninstallation

 1. To destroy all your running servers, run `terraform destroy` inside the
    `/terraform` directory and follow the on-screen instructions.

Terraform will prompt you to enter variable values upon destruction, however
you can enter any non-empty value.

In case of failure, manually destroy the servers from DigitalOcean and remove
the `terraform.tfstate` and `terraform.tfstate.backup` files in the `/terraform`
directory.

## GitHub integration

Refer to the README inside the `/loadtests.me` directory for more information.