resource "hrui_loop_protocol" "protocol_off" {
  loop_function = "Off"
}

resource "hrui_loop_protocol" "protocol_detection" {
  loop_function = "Loop Detection"
  interval_time = 30
  recover_time  = 15
}

resource "hrui_loop_protocol" "protocol_stp" {
  loop_function = "Spanning Tree"
}
