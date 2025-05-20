resource "google_compute_global_address" "default" {
  name = "${var.name}-address"
}

resource "google_compute_health_check" "default" {
  name               = "${var.name}-health-check"
  timeout_sec        = 5
  check_interval_sec = 10

  tcp_health_check {
    port = 50051
  }
}

resource "google_compute_backend_service" "default" {
  name                  = "${var.name}-backend-service"
  protocol              = "HTTP"
  port_name             = "grpc"
  timeout_sec           = 30
  health_checks         = [google_compute_health_check.default.id]
  load_balancing_scheme = "EXTERNAL_MANAGED"

  dynamic "backend" {
    for_each = var.backend_services
    content {
      group           = backend.value.group
      balancing_mode  = "UTILIZATION"
      capacity_scaler = 1.0
    }
  }
}

resource "google_compute_url_map" "default" {
  name            = "${var.name}-url-map"
  default_service = google_compute_backend_service.default.id
}

resource "google_compute_target_http_proxy" "default" {
  name    = "${var.name}-http-proxy"
  url_map = google_compute_url_map.default.id
}

resource "google_compute_global_forwarding_rule" "default" {
  name                  = "${var.name}-forwarding-rule"
  target                = google_compute_target_http_proxy.default.id
  port_range            = "80"
  ip_address            = google_compute_global_address.default.address
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

output "ip_address" {
  value = google_compute_global_address.default.address
}
