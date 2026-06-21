import keycloak from "#/lib/auth/keycloak";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/(protected)")({
  component: RouteComponent,
  beforeLoad: async () => {
    if (!keycloak.authenticated) {
      await keycloak.init({
        onLoad: "login-required",
        redirectUri: `${location.origin}/home`,
      });
    }
    if (!keycloak.authenticated) {
      throw redirect({
        to: "/",
      });
    }
  },
});

function RouteComponent() {
  return <Outlet />;
}
