# Example usage of stp_global_settings resource
resource "hrui_stp_global_settings" "example" {
  force_version      = "RSTP"                       # Options: "STP", "RSTP"
  priority           = 32768                        # Bridge priority (default 32768)
  max_age            = 20                           # Time (in seconds) a bridge waits before forgetting topology info
  hello_time         = 2                            # Time interval (in seconds) between configuration messages
  forward_delay      = 15                           # Delay (in seconds) for transitioning port states
}
