data "hrui_system_info" "switch" {}

output "system_info" {
  value = data.hrui_system_info.switch
}

