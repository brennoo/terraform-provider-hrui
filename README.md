# Terraform Provider for HRUI Switches

This is a hobby project to create a Terraform provider for [HRUI](www.hruitech.com) switches that are web-managed. It relies on [goquery](https://github.com/PuerkitoBio/goquery) for interfacing with the web UI.

Check the documentation at:

- Terraform: [HRUI Provider](https://registry.terraform.io/providers/brennoo/hrui)
- OpenToFu: [HRUI Provider](https://search.opentofu.org/provider/brennoo/hrui)

## HRUI ODM/OEM

This provider is developed using a Horaco (ZX-SWTG124AS) switch. Other brands that are likely to work with this provider, as they seem to be the same hardware, include:

* HRUI
* Horaco
* Sodola
* XikeStor
* AmpCom

> [!NOTE]
> ## Using firmware v1.9.

## Getting Started

### Requirements

* Go 1.23 or later
* Terraform 1.9 or later

### Building the provider

1. Clone the repository:

    ```bash
    git clone github.com/brennoo/terraform-provider-hrui.git
    ```

2. Navigate to the provider directory:

    ```bash
    cd terraform-provider-hrui
    ```

3. Build the provider:

    ```bash
    make build
    ```

    This will create an executable file named `terraform-provider-hrui` in the project directory.

### Using the provider

1.  Configure the provider in your Terraform configuration:

    ```terraform
    terraform {
      required_providers {
        hrui = {
          source  = "brennoo/hrui"
          version = "= 0.1.0-alpha.1"
        }
      }
    }

    provider "hrui" {
      url      = "http://192.168.2.1"
      username = "admin"
      # The password can be set in HRUI_PASSWORD environment variable
    }
    ```

2.  Refer to the [examples](examples) folder for usage of available resources and data sources.

### Developing the provider

1.  Set up a development override in your `~/.terraformrc` file:

    ```
    provider_installation {
      dev_overrides {
        "brennoo/hrui" = "/path/to/your/project"
      }
      direct {}
    }
    ```

    Replace `/path/to/your/project` with the actual path to your provider directory.

2.  Run `terraform init` in your Terraform project directory.

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines.

## License

This provider is licensed under the MPL-2.0 License. See the [LICENSE](LICENSE) file for details.

