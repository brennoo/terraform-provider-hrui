# Contributing to terraform-provider-hrui

Contributions are welcome! Here's how you can contribute to this project:
### Reporting bugs

If you find a bug, please [open an issue](https://github.com/brennoo/terraform-provider-hrui/issues/new) on the GitHub repository. Provide as much detail as possible, including:

*   Steps to reproduce the bug
*   Expected behavior
*   Actual behavior
*   Terraform version
*   Provider version
*   Switch model and firmware version
*   The HTML page of the concerned resource

### Suggesting enhancement

If you have an idea for an enhancement or a new feature, please [open an issue](https://github.com/brennoo/terraform-provider-hrui/issues/new) on the GitHub repository. Describe the enhancement and explain why it would be beneficial.

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


### Local override
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

### Submitting pull requests

If you want to contribute code, please follow these steps:

1.  Fork the repository.
2.  Create a new branch for your changes.
3.  Make your changes and commit them with descriptive commit messages.
4.  Push your changes to your fork.  

5.  Open a pull request :tada:

Please ensure that your code follows the existing code style and includes tests for any new functionality.
