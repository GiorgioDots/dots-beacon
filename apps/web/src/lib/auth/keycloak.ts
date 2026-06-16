import Keycloak from "keycloak-js";

const keycloak = new Keycloak({
  url: "http://localhost:8081",
  realm: "dots-beacon",
  clientId: "dots-beacon-app",
});

export default keycloak;
