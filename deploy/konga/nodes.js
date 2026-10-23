// Konga seed: the Kong connection. Konga calls Kong from inside Docker, so the
// Admin URL uses the compose service name, not localhost.
module.exports = [
  {
    name: "Chura Kong",
    type: "default",
    kong_admin_url: "http://kong:8001",
    health_checks: false,
  },
];
