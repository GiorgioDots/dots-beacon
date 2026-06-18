import { Button } from "#/components/ui/button";
import keycloak from "#/lib/auth/keycloak";
import { createFileRoute, Link } from "@tanstack/react-router";

export const Route = createFileRoute("/(protected)/home/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <>
      <Button
        onClick={() => keycloak.logout({ redirectUri: `${location.origin}` })}
      >
        logout
      </Button>
      <Link to="/">Index</Link>
    </>
  );
}
