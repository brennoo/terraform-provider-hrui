terraform {
  required_providers {
    hrui = {
      source  = "brennoo/hrui"
      version = "~> 0.1.0"
    }
  }
}

provider "hrui" {
  url      = "http://192.168.2.1" # Replace with the URL of your HRUI switch
  username = "admin"              # Replace with the username for authentication
  password = "password"           # Replace with the password for authentication
  autosave = true                 # Save all changes every terraform apply, default true
}

# These configurations are also available as environment variables HRUI_XXX


