resource "hrui_stp_global_settings" "example" {
  force_version = "RSTP"
  priority      = 32768
  max_age       = 20
  hello_time    = 2
  forward_delay = 15
}
